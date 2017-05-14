package agent

import (
	"fmt"
	"net/url"
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/cmdproxy/pkg/test"
)

func TestQueryParams(t *testing.T) {
	t.Parallel()

	t.Run("decode", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			var (
				qp QueryParams

				s      = a.String()
				u, err = url.Parse(fmt.Sprintf("http://example.com?info=%s", s))
			)
			if err != nil {
				t.Error(err)
			}
			if err := qp.DecodeFrom(u, queryRequired); err != nil {
				t.Error(err)
			}

			return qp.Info == s
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("decode required", func(t *testing.T) {
		var (
			qp     QueryParams
			u, err = url.Parse("http://example.com")
		)
		if err != nil {
			t.Error(err)
		}
		if err := qp.DecodeFrom(u, queryRequired); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("decode optional", func(t *testing.T) {
		var (
			qp     QueryParams
			u, err = url.Parse("http://example.com")
		)
		if err != nil {
			t.Error(err)
		}
		if err := qp.DecodeFrom(u, queryOptional); err != nil {
			t.Errorf("expected error")
		}
	})
}
