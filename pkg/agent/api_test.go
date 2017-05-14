package agent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/cmdproxy/pkg/test"
	"github.com/go-kit/kit/log"
)

func TestAPI(t *testing.T) {
	t.Parallel()

	var (
		api    = NewAPI(0, log.NewNopLogger())
		server = httptest.NewServer(api)
		url    = server.URL
	)

	t.Run("update", func(t *testing.T) {
		fn := func(s test.ASCII) bool {
			resp, err := http.Get(fmt.Sprintf("%s/update?info=%s", url, s.String()))
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
