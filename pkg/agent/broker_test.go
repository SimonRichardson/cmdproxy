package agent

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

func TestBrokers(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()

	t.Run("add", func(t *testing.T) {
		var (
			broker    = NewBroker(nil, "tcp", "0.0.0.0:0", logger)
			brokers   = NewBrokers(logger)
			addr, err = brokers.Add(broker)
		)
		if err != nil {
			t.Error(err)
		}
		if addr == "" {
			t.Errorf("expected: valid address, actual: %s", addr)
		}

		_, port, err := net.SplitHostPort(addr)
		if err != nil {
			t.Error(err)
		}

		if _, err := strconv.Atoi(port); err != nil {
			t.Error(err)
		}
	})

	t.Run("serve", func(t *testing.T) {
		var (
			api       = NewAPI(0, logger)
			broker    = NewBroker(api, "tcp", "0.0.0.0:0", logger)
			brokers   = NewBrokers(logger)
			addr, err = brokers.Add(broker)
		)
		if err != nil {
			t.Error(err)
		}
		_, port, err := net.SplitHostPort(addr)
		if err != nil {
			t.Error(err)
		}

		go func() {
			if e := brokers.Serve(); e != nil {
				t.Error(e)
			}
			defer brokers.Close()
		}()

		var (
			wg   sync.WaitGroup
			errs = make(chan error)
		)
		wg.Add(1)

		go func() {
			defer wg.Done()

			res, err := http.Get(fmt.Sprintf("http://0.0.0.0:%s/update?info=hello", port))
			if err != nil {
				errs <- err
				return
			}
			if res.StatusCode != http.StatusOK {
				errs <- errors.New("invalid response")
			}
		}()

		go func() { wg.Wait(); close(errs) }()

		for e := range errs {
			if e != nil {
				t.Error(err)
			}
		}
	})
}
