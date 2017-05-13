package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"strings"

	"github.com/SimonRichardson/cmdproxy/pkg/group"
	"github.com/SimonRichardson/cmdproxy/pkg/peer"
	"github.com/SimonRichardson/cmdproxy/pkg/proxy"
	"github.com/SimonRichardson/cmdproxy/pkg/scheduler"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

// runForward manages all the state between the various agents
func runForward(args []string) error {
	var (
		flagset = flag.NewFlagSet("forward", flag.ExitOnError)

		debug   = flagset.Bool("debug", false, "debug logging")
		apiAddr = flagset.String("api", defaultAPIAddr, "listen address for store API")

		agents = stringSlice{}
	)
	flagset.Var(&agents, "agents", "agent host host:peer (repeatable)")
	flagset.Usage = usageFor(flagset, "forward [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}
	// Forward can be used in a pipe constructor.
	// example: `proxy agents -agents.broker-size=3 | proxy forward`
	if info, err := os.Stdin.Stat(); err == nil && (info.Mode()&os.ModeCharDevice) != os.ModeCharDevice {
		// Read from the Stdin for any possible pipe arguments.
		var (
			reader    = bufio.NewReader(os.Stdin)
			line, err = reader.ReadString('\n')
		)
		if err != nil || (len(args) == 0 && len(line) == 0) {
			return errors.New("specify at least one agent address via pipe")
		}

		parts := strings.Split(strings.TrimSpace(line), " ")
		if err := flagset.Parse(append(args, parts...)); err != nil {
			return err
		}
	} else if flagset.NFlag() == 0 {
		// Nothing found in pipe, which means nothing via flags.
		return errors.New("specify at least one agent address as an argument")
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

	level.Info(logger).Log("Agents", agents.String())

	// Parse URLs for listeners.
	apiNetwork, apiAddress, err := parseAddr(*apiAddr, defaultAPIPort)
	if err != nil {
		return err
	}

	// Create a client that is reusable by the peer clients.
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			ResponseHeaderTimeout: 5 * time.Second,
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 1,
		},
	}

	// Create proxy peer agents.
	peers := make([]*peer.Peer, agents.Len())
	for i := 0; i < agents.Len(); i++ {
		_, agentAddr, err := parseAddr(agents[i], defaultAgentAPIPort)
		if err != nil {
			return err
		}

		peers[i] = peer.NewPeer(
			client,
			"http",
			agentAddr,
			log.With(logger, "component", "peer"),
		)
	}

	// Scheduler runs tasks on the proxy peer agents.
	scheduler := scheduler.NewScheduler(
		peers,
		log.With(logger, "component", "scheduler"),
	)

	// Bind listeners.
	apiListener, err := net.Listen(apiNetwork, apiAddress)
	if err != nil {
		return err
	}
	level.Info(logger).Log("API", fmt.Sprintf("%s://%s", apiNetwork, apiAddress))

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
		// Set up the scheduler for tasks to be worked on.
		g.Add(func() error {
			scheduler.Run()
			return nil
		}, func(error) {
			scheduler.Stop()
		})
	}
	{
		// Set up the new server mux
		g.Add(func() error {
			var (
				mux = http.NewServeMux()
				api = proxy.NewAPI(
					scheduler,
					logger,
				)
			)
			defer func() {
				if err := api.Close(); err != nil {
					level.Warn(logger).Log("err", err)
				}
			}()

			mux.Handle("/proxy/", http.StripPrefix("/proxy", api))

			return http.Serve(apiListener, mux)
		}, func(error) {
			apiListener.Close()
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
