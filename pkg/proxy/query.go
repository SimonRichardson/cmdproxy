package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// RunQueryParams defines all the dimensions of a query.
type RunQueryParams struct {
	ClientID    int    `json:"client_id"`
	Info        string `json:"info"`
	Mode        string `json:"mode"`
	FailOnError bool   `json:"failonerror"`
}

// DecodeFrom populates a RunQueryParams from a URL.
func (qp *RunQueryParams) DecodeFrom(u *url.URL, rb queryBehaviour) error {
	// Required depending on the query behaviour
	var (
		err      error
		clientID = u.Query().Get("client_id")
	)
	qp.ClientID, err = strconv.Atoi(clientID)
	if err != nil || (clientID == "" && rb == queryRequired) {
		return errors.New("Error reading/parsing 'client_id' (required) query.")
	}

	qp.Info = u.Query().Get("info")
	if qp.Info == "" && rb == queryRequired {
		return errors.New("Error reading/parsing 'info' (required) query.")
	}

	qp.Mode = u.Query().Get("mode")
	if qp.Mode == "" && rb == queryRequired {
		return errors.New("Error reading/parsing 'mode' (required) query.")
	}

	// Optional
	failOnError := u.Query().Get("failonerror")
	if qp.FailOnError, err = strconv.ParseBool(failOnError); err != nil {
		return errors.New("Error reading/parsing 'failOnError' (required) query.")
	}

	return nil
}

// RunQueryResult contains statistics about the query.
type RunQueryResult struct {
	Params   RunQueryParams `json:"query"`
	Duration string         `json:"duration"`

	Records string
}

// EncodeTo encodes the QueryResult to the HTTP response writer.
func (qr *RunQueryResult) EncodeTo(w http.ResponseWriter) {
	w.Header().Set(httpHeaderClientID, strconv.Itoa(qr.Params.ClientID))
	w.Header().Set(httpHeaderInfo, qr.Params.Info)
	w.Header().Set(httpHeaderMode, qr.Params.Mode)
	w.Header().Set(httpHeaderFailOnError, strconv.FormatBool(qr.Params.FailOnError))
	w.Header().Set(httpHeaderDuration, qr.Duration)

	fmt.Fprintf(w, qr.Records)
}

// QueryParams defines all the dimensions of a query.
type QueryParams struct {
	TaskID string `json:"task_id"`
}

// DecodeFrom populates a QueryParams from a URL.
func (qp *QueryParams) DecodeFrom(u *url.URL, rb queryBehaviour) error {
	// Required depending on the query behaviour
	qp.TaskID = u.Query().Get("task_id")
	if qp.TaskID == "" && rb == queryRequired {
		return errors.New("Error reading/parsing 'info' (required) query.")
	}

	return nil
}

// QueryResult contains statistics about the query.
type QueryResult struct {
	Params   QueryParams `json:"query"`
	Duration string      `json:"duration"`

	Records string
}

// EncodeTo encodes the QueryResult to the HTTP response writer.
func (qr *QueryResult) EncodeTo(w http.ResponseWriter) {
	w.Header().Set(httpHeaderTaskID, qr.Params.TaskID)
	w.Header().Set(httpHeaderDuration, qr.Duration)

	fmt.Fprintf(w, qr.Records)
}

const (
	httpHeaderClientID    = "X-Proxy-ClientID"
	httpHeaderInfo        = "X-Proxy-Info"
	httpHeaderMode        = "X-Proxy-Mode"
	httpHeaderFailOnError = "X-Proxy-FailOnError"
	httpHeaderTaskID      = "X-Proxy-TaskID"
	httpHeaderDuration    = "X-Proxy-Duration"
)

type queryBehaviour int

const (
	queryRequired queryBehaviour = iota
	queryOptional
)
