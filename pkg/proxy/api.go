package proxy

import (
	"net/http"
	"time"

	"github.com/SimonRichardson/cmdproxy/pkg/scheduler"
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
	scheduler *scheduler.Scheduler
	logger    log.Logger
}

// NewAPI creates a API with the correct dependencies.
func NewAPI(scheduler *scheduler.Scheduler, logger log.Logger) *API {
	return &API{
		scheduler: scheduler,
		logger:    logger,
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
	// useful metrics
	begin := time.Now()

	// Valdiate user input.
	var qp RunQueryParams
	if err := qp.DecodeFrom(r.URL, queryRequired); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if qp.ClientID < 0 || qp.ClientID >= len(a.scheduler.Peers()) {
		http.Error(w, "Invalid client ID.", http.StatusBadRequest)
		return
	}

	modeType, err := scheduler.ParseModeType(qp.Mode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Write the information to the peers.
	task := scheduler.NewTask(modeType, qp.ClientID, qp.Info, qp.FailOnError)
	a.scheduler.Register(task)

	// We'll collect responese into a single RunQueryResult
	qr := RunQueryResult{Params: qp}
	qr.Records = task.ID()

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func (a *API) handleStatusQuery(w http.ResponseWriter, r *http.Request) {
	// useful metrics
	begin := time.Now()

	// Valdiate user input.
	var qp QueryParams
	if err := qp.DecodeFrom(r.URL, queryRequired); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, ok := a.scheduler.Get(qp.TaskID)
	if !ok {
		http.Error(w, "no task found", http.StatusNotFound)
		return
	}

	// We'll collect responese into a single QueryResult
	qr := QueryResult{Params: qp}
	qr.Records = string(task.Status())

	// Finish
	qr.Duration = time.Since(begin).String()
	qr.EncodeTo(w)
}

func (a *API) handleStopQuery(w http.ResponseWriter, r *http.Request) {
	// useful metrics
	begin := time.Now()

	// Valdiate user input.
	var qp QueryParams
	if err := qp.DecodeFrom(r.URL, queryRequired); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, ok := a.scheduler.Get(qp.TaskID)
	if !ok {
		http.Error(w, "no task found", http.StatusNotFound)
		return
	}

	// Attempt to cancel the task.
	a.scheduler.Cancel(task)

	// We'll collect responese into a single QueryResult
	qr := QueryResult{Params: qp}
	qr.Records = string(task.Status())

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
