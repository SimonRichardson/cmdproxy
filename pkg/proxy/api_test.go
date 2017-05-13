package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/quick"

	"io/ioutil"

	"github.com/SimonRichardson/cmdproxy/pkg/peer"
	"github.com/SimonRichardson/cmdproxy/pkg/scheduler"
	"github.com/SimonRichardson/cmdproxy/pkg/test"
	"github.com/go-kit/kit/log"
)

func TestAPI(t *testing.T) {
	var (
		logger    = log.NewNopLogger()
		scheduler = scheduler.NewScheduler([]*peer.Peer{
			peer.NewPeer(http.DefaultClient, "tcp", "0.0.0.0:0", logger),
		}, logger)
		api    = NewAPI(scheduler, logger)
		server = httptest.NewServer(api)
		url    = server.URL
	)

	t.Run("run", func(t *testing.T) {
		fn := func(s test.ASCII) bool {
			resp, err := http.Get(fmt.Sprintf("%s/run?client_id=0&info=%s&mode=parallel&failonerror=false", url, s.String()))
			if err != nil {
				t.Error(err)
			}

			return resp.StatusCode == http.StatusOK
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("status", func(t *testing.T) {
		fn := func(s test.ASCII) bool {
			resp, err := http.Get(fmt.Sprintf("%s/run?client_id=0&info=%s&mode=parallel&failonerror=false", url, s.String()))
			if err != nil {
				t.Error(err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected: %v, actual: %v", http.StatusOK, resp.StatusCode)
			}

			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
			}

			taskID := string(bytes)
			resp, err = http.Get(fmt.Sprintf("%s/status?task_id=%s", url, taskID))
			if err != nil {
				t.Error(err)
			}

			return resp.StatusCode == http.StatusOK
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("stop", func(t *testing.T) {
		fn := func(s test.ASCII) bool {
			resp, err := http.Get(fmt.Sprintf("%s/run?client_id=0&info=%s&mode=parallel&failonerror=false", url, s.String()))
			if err != nil {
				t.Error(err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected: %v, actual: %v", http.StatusOK, resp.StatusCode)
			}

			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
			}

			taskID := string(bytes)
			resp, err = http.Get(fmt.Sprintf("%s/stop?task_id=%s", url, taskID))
			if err != nil {
				t.Error(err)
			}

			return resp.StatusCode == http.StatusOK
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/bad", url))
		if err != nil {
			t.Error(err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected: %v, actual: %v", http.StatusNotFound, resp.StatusCode)
		}
	})
}
