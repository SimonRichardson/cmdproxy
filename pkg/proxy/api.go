package proxy

import (
	"net/http"

	"github.com/go-kit/kit/log"
)

// These are the proxy API URL paths.
const (
	APIPathRunQuery    = "/run"
	APIPathStatusQuery = "/status"
	APIPathStopQuery   = "/stop"
)

// API serves the proxy API
type API struct {
	logger log.Logger
}

func NewAPI(logger log.Logger) *API {
	return &API{
		logger: logger,
	}
}

// Close out the API
func (a *API) Close() error {
	return nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Routing table
	method, path := r.Method, r.URL.Path
	switch {
	case method == "GET" && path == APIPathRunQuery:
		a.handleRunQuery(w, r)
	case method == "GET" && path == APIPathStatusQuery:
		a.handleStatusQuery(w, r)
	case method == "GET" && path == APIPathStopQuery:
		a.handleStopQuery(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *API) handleRunQuery(w http.ResponseWriter, r *http.Request) {

}

func (a *API) handleStatusQuery(w http.ResponseWriter, r *http.Request) {

}

func (a *API) handleStopQuery(w http.ResponseWriter, r *http.Request) {

}
