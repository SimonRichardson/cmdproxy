package agent

import (
	"fmt"
	"net"

	"net/http"

	"github.com/SimonRichardson/cmdproxy/pkg/group"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Broker is a way to manage it's own API
type Broker struct {
	api                    *API
	apiNetwork, apiAddress string
	apiListener            net.Listener
	logger                 log.Logger
}

// NewBroker creates a new broker that well self manage it's own API.
func NewBroker(api *API, apiNetwork, apiAddress string, logger log.Logger) *Broker {
	return &Broker{
		api:        api,
		apiNetwork: apiNetwork,
		apiAddress: apiAddress,
		logger:     logger,
	}
}

// Bind the API address for the broker
func (b *Broker) Bind() (addr string, err error) {
	// Bind listeners
	b.apiListener, err = net.Listen(b.apiNetwork, b.apiAddress)
	if err != nil {
		return addr, err
	}

	address := b.apiListener.Addr().String()
	level.Debug(b.logger).Log("broker-addr", address)
	return address, nil
}

// Serve API via the Broker
func (b *Broker) Serve() error {
	mux := http.NewServeMux()
	mux.Handle("/", b.api)

	// Help with debugging
	level.Debug(b.logger).Log("serving", fmt.Sprintf("%s%s", b.apiListener.Addr().String(), "/"))

	return http.Serve(b.apiListener, mux)
}

// Close the API on the Broker
func (b *Broker) Close() error {
	if err := b.api.Close(); err != nil {
		level.Warn(b.logger).Log("component", "API", "err", err)
	}

	if err := b.apiListener.Close(); err != nil {
		level.Warn(b.logger).Log("component", "API Listener", "err", err)
		return err
	}

	return nil
}

// Brokers manages a set of brokers
type Brokers struct {
	group  group.Group
	stop   chan struct{}
	logger log.Logger
}

// NewBrokers creates a new Broker manager
func NewBrokers(logger log.Logger) *Brokers {
	var (
		g    group.Group
		stop = make(chan struct{})
	)
	{
		// Make sure that we close everything we're executing.
		g.Add(func() error {
			select {
			case <-stop:
				return nil
			}
		}, func(error) {
			close(stop)
		})
	}
	return &Brokers{
		group:  g,
		stop:   stop,
		logger: logger,
	}
}

// Add a broker to manage.
func (b *Brokers) Add(broker *Broker) (addr string, err error) {
	addr, err = broker.Bind()
	if err != nil {
		return addr, err
	}

	b.group.Add(func() error {
		return broker.Serve()
	}, func(error) {
		broker.Close()
	})

	return addr, err
}

// Serve the managed brokers.
func (b *Brokers) Serve() error {
	return b.group.Run()
}

// Close the managed brokers.
func (b *Brokers) Close() {
	b.stop <- struct{}{}
}
