# cmdproxy

Command Proxy

- [Getting started](#example)
- [Introduction](#introduction)
- [Agents](#agents)
- [Proxy](#proxy)
- [Improvements](#improvements)

## Example

Quick guide to getting started.

```
make all
cd dist

proxy agents | proxy forward
```

*NOTE:* The above command is using the "power user" mode, as it is tiring copying and
pasting agents urls to the proxy CLI.

## Introduction

Command Proxy is a cli for wrapping a set of distributed workers (agents).
See the diagram bellow to better understand the architecture:

```
+---------+   +---------+   +---------+
|         |   |         |   |         |
|  Agent  |   |  Agent  |   |  Agent  |
|         |   |         |   |         |
+----^----+   +----^----+   +----^----+
     |             |             |
     +-------------+-------------+
                   |
            +------+------+
            |             |
            |  Async API  |
            |             |
            +-------------+
```

### Agents

To help better test the command proxy cli (see bellow), the cli can generate
any number of agents that can be fed into the proxy forwarding cli for it to
use.

#### Agent REST API

The agent REST API has only one route and that's the following:

 - `/update` - takes one parameter `info`, which has to be a string and returns
 http StatusOK if successful or a `plain/text` error on failure.

#### Agent CLI API

In order to generate a number of agents, you can use the following command
(assuming you've compiled the project).

```
proxy agents
```

It will by default create 3 proxy agents for you, but can be overridden using
`-agents.broker-size=5`. In this example it will generate 5 broker agents. Also
available to you is a `-help` parameter that list all the various options to
help better test and understand how to make those agents.

```
proxy agents -help
USAGE
  forward [flags]

FLAGS
  -agents.api tcp://0.0.0.0:0  listen address for agenet API
  -agents.broker-size 3        amount of agent brokers required
  -debug false                 debug logging
  -delay 5m0s                  delay duration to make agents more realistic
  -output.addresses true       output addresses defines if agents url should be forwarded to stdout
  -output.prefix -agents       output prefix defines what prefixes should be used for output.addresses
```

### Proxy

Proxy command takes a series of agent urls and allows a series of REST API
calls to work with the agents (see above).

#### Proxy REST API

The proxy REST API has three routes:

 - `start` - takes four parameters and returns http StatusOK and a task ID if
 the request is successful or a `plain/text` error on failure.
    - `client_id` - defines the offset at which agent to start the requests with.
    - `info` - defines what to send to the agents
    - `failonerror` - defines if work should continue when a request errors out.
    - `mode` - defines if a request should be parallel or sequential.
 - `status` - takes only one parameter and returns a `plain/text` string of the
 status of the task or error on failure.
    - `task_id` - defines which task you'd like to know the status of.
 - `kill` - takes only one parameter and returns http StatusOK or `plain/text`
 error on failure.

#### Proxy CLI API

The proxy CLI API takes a series of agent urls for it to send the `info`
parameter to. In order to make life more productive, you can use the CLI in
power user mode, by getting the above agents CLI API to feed to the `stdin` of
the proxy CLI API.

```
proxy agents | proxy forward
```

The above command will generate 3 agents and output 3 urls for the `proxy forward`
command to consume on the stdin. Alternatively you can do the following:

```
# In one session call this, but wait for the stdout of urls
proxy agents -output.prefix=""
# [::]:49834 [::]:49835 [::]:49836
```

```
# In another session, copy and paste the 3 output urls to the following command
proxy [::]:49834 [::]:49835 [::]:49836
```

## Improvements

Possible improvements:

 - Better support for agents, including use of something like coordination free
 member management i.e. hashicorp/serf or hashicorp/memberslist
 - Storing the tasks in a KVS so that the proxy REST API can it self be
 distributed. That way any scheduler can from any proxy can work on tasks. This
 way if a proxy REST API goes down, the tasks can still be processed.
