package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/SimonRichardson/cmdproxy/pkg/group"
	"github.com/SimonRichardson/cmdproxy/pkg/proxy"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

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
	args = flagset.Args()
	if len(args) <= 0 {
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

	// Parse URLs for listeners.
	apiNetwork, apiAddress, _, _, err := parseAddr(*apiAddr, defaultAPIPort)
	if err != nil {
		return err
	}

	// Bind listeners.
	apiListener, err := net.Listen(apiNetwork, apiAddress)
	if err != nil {
		return err
	}
	level.Info(logger).Log("API", fmt.Sprintf("%s://%s", apiNetwork, apiAddress))

	// Execution group.
	var g group.Group
	{
		// Make sure that we close everything we're executing.
		cancel := make(chan struct{})
		g.Add(func() error {
			<-cancel
			return nil
		}, func(error) {
			close(cancel)
		})
	}
	{
		// Set up the new server mux
		g.Add(func() error {
			var (
				mux = http.NewServeMux()
				api = proxy.NewAPI(logger)
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
