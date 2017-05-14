package peer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/cmdproxy/pkg/test"
	"github.com/go-kit/kit/log"
)

func TestPeer(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()

	t.Run("newrequest", func(t *testing.T) {
		fn := func(s test.ASCII) bool {
			peer := NewPeer(http.DefaultClient, "http", "0.0.0.0:0", logger)
			if _, err := peer.NewRequest(s.String()); err != nil {
				return false
			}
			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestRequest(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()

	t.Run("do", func(t *testing.T) {
		fn := func(s test.ASCII) bool {
			var (
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				addr     = strings.Replace(server.URL, "http://", "", 1)
				peer     = NewPeer(http.DefaultClient, "http", addr, logger)
				req, err = peer.NewRequest(s.String())
			)
			if err != nil {
				return false
			}

			resp, err := req.Do()
			if err != nil {
				return false
			}

			return resp.StatusCode == http.StatusOK
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("url", func(t *testing.T) {
		fn := func(s test.ASCII) bool {
			var (
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				addr     = strings.Replace(server.URL, "http://", "", 1)
				peer     = NewPeer(http.DefaultClient, "http", addr, logger)
				req, err = peer.NewRequest(s.String())
			)
			if err != nil {
				return false
			}

			return fmt.Sprintf("%s/update?info=%s", server.URL, s.String()) == req.URL().String()
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}
