package agent

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// QueryParams defines all dimensions of a query.
type QueryParams struct {
	Info string `json:"info"`
}

// DecodeFrom populates a QueryParams from a URL.
func (qp *QueryParams) DecodeFrom(u *url.URL, rb queryBehaviour) error {
	// Required depending on the query behaviour
	qp.Info = u.Query().Get("info")
	if qp.Info == "" && rb == queryRequired {
		return errors.New("Error reading/parsing 'info' (required) query.")
	}

	return nil
}

type queryBehaviour int

const (
	queryRequired queryBehaviour = iota
	queryOptional
)

// QueryResult contains statistics about the query.
type QueryResult struct {
	Params   QueryParams `json:"query"`
	Duration string      `json:"duration"`
}

// EncodeTo encodes the QueryResult to the HTTP response writer.
func (qr *QueryResult) EncodeTo(w http.ResponseWriter) {
	w.Header().Set(httpHeaderInfo, qr.Params.Info)
	w.Header().Set(httpHeaderDuration, qr.Duration)
}

const (
	httpHeaderInfo     = "X-Proxy-Info"
	httpHeaderDuration = "X-Proxy-Duration"
)
