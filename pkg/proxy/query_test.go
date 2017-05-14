package proxy

import (
	"fmt"
	"net/url"
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/cmdproxy/pkg/test"
)

func TestRunQueryParams(t *testing.T) {
	t.Parallel()

	t.Run("decode", func(t *testing.T) {
		fn := func(a uint, b, c test.ASCII) bool {
			var (
				qp RunQueryParams

				id, i, m = a % 9999, b.String(), c.String()
				u, err   = url.Parse(fmt.Sprintf("http://example.com?client_id=%d&info=%s&mode=%s", id, i, m))
			)
			if err != nil {
				t.Error(err)
			}
			if err := qp.DecodeFrom(u, queryRequired); err != nil {
				t.Error(err)
			}

			return qp.ClientID == int(id) && qp.Info == i && qp.Mode == m
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("decode required", func(t *testing.T) {
		var (
			qp     RunQueryParams
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
			qp     RunQueryParams
			u, err = url.Parse("http://example.com?client_id=0&info=hello&mode=world")
		)
		if err != nil {
			t.Error(err)
		}
		if err := qp.DecodeFrom(u, queryOptional); err != nil {
			t.Errorf("expected error")
		}
	})
}
