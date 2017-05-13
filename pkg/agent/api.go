package agent

import (
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// These are the agent API URL paths.
const (
	APIPathUpdateQuery = "/update"
)

// API serves the agent API
type API struct {
	delay  time.Duration
	logger log.Logger
}

func NewAPI(delay time.Duration, logger log.Logger) *API {
	return &API{
		delay:  delay,
		logger: logger,
	}
}

// Close out the API
func (a *API) Close() error {
	return nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	iw := &interceptingWriter{http.StatusOK, w}
	w = iw
	// Routing table
	method, path := r.Method, r.URL.Path
	switch {
	case method == "GET" && path == APIPathUpdateQuery:
		a.handleUpdateQuery(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *API) handleUpdateQuery(w http.ResponseWriter, r *http.Request) {
	// useful metrics
	begin := time.Now()

	// Valdiate user input.
	var qp QueryParams
	if err := qp.DecodeFrom(r.URL, queryRequired); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	level.Debug(a.logger).Log("info", qp.Info, "delay", a.delay.String())

	// We'll collect responese into a single QueryResult
	qr := QueryResult{Params: qp}

	// Sleep for sometime, just to make it feel more relistic.
	if a.delay > 0 {
		time.Sleep(a.delay)
	}

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

type interceptingWriter struct {
	code int
	http.ResponseWriter
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}
