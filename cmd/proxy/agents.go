package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SimonRichardson/cmdproxy/pkg/agent"
	"github.com/SimonRichardson/cmdproxy/pkg/group"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	defaultDelay            = time.Second * 3
	defaultAgentBrokersSize = 3
	defaultOutputAddresses  = true
	defaultOutputPrefix     = "-agents"
	defaultForwardAPIPort   = 0
)

var (
	defaultAgentAPIAddr = fmt.Sprintf("tcp://0.0.0.0:%d", defaultForwardAPIPort)
)

// runAgents manages all the agents
func runAgents(args []string) error {
	var (
		flagset = flag.NewFlagSet("agents", flag.ExitOnError)

		debug = flagset.Bool("debug", false, "debug logging")
		delay = flagset.Duration("delay", defaultDelay, "delay duration to make agents more realistic")

		outputAddresses = flagset.Bool("output.addresses", defaultOutputAddresses, "output addresses defines if agents url should be forwarded to stdout")
		outputPrefix    = flagset.String("output.prefix", defaultOutputPrefix, "output prefix defines what prefixes should be used for output.addresses")

		agentAPIAddr    = flagset.String("agents.api", defaultAgentAPIAddr, "listen address for agenet API")
		agentBrokerSize = flagset.Int("agents.broker-size", defaultAgentBrokersSize, "amount of agent brokers required")
	)
	flagset.Usage = usageFor(flagset, "forward [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	// Logging.
	var logger log.Logger
	{
		logLevel := level.AllowInfo()
		if *debug {
			logLevel = level.AllowAll()
		}
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = level.NewFilter(logger, logLevel)
	}

	apiNetwork, apiAddress, err := parseAddr(*agentAPIAddr, defaultForwardAPIPort)
	if err != nil {
		return err
	}

	var (
		brokers = agent.NewBrokers(log.With(logger, "component", "brokers"))
		addrs   = make([]string, *agentBrokerSize)
	)
	for i := 0; i < *agentBrokerSize; i++ {
		addr, err := brokers.Add(agent.NewBroker(
			agent.NewAPI(
				*delay,
				logger,
			),
			apiNetwork,
			apiAddress,
			log.With(logger, "component", "broker"),
		))
		if err != nil {
			return err
		}
		addrs[i] = addr
	}
	level.Debug(logger).Log("addrs", strings.Join(addrs, ", "))

	// Execution group.
	var g group.Group
	{
		cancel := make(chan struct{})
		g.Add(func() error {
			<-cancel
			return nil
		}, func(error) {
			close(cancel)
		})
	}
	{
		g.Add(func() error {
			output := make([]string, len(addrs))
			for i, v := range addrs {
				if fix := *outputPrefix; fix != "" {
					output[i] = fmt.Sprintf("%s=", fix)
				}
				output[i] += v
			}

			if *outputAddresses {
				fmt.Fprintln(os.Stdout, strings.Join(output, " "))
			}

			return brokers.Serve()
		}, func(error) {
			fmt.Println("CLOSE")
			brokers.Close()
		})
	}
	{
		// Setup os signal interruptions.
		cancel := make(chan struct{})
		g.Add(func() error {
			return interrupt(cancel)
		}, func(error) {
			close(cancel)
		})
	}
	return g.Run()
}
