package peer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
)

// Peer wraps a proxy for requesting and cancelling.
type Peer struct {
	client  *http.Client
	network string
	addr    string
	logger  log.Logger
}

// NewPeer creates a new peer using a resuable client.
func NewPeer(client *http.Client, network, addr string, logger log.Logger) *Peer {
	return &Peer{
		client:  client,
		network: network,
		addr:    addr,
		logger:  logger,
	}
}

// NewRequest creates a new request ready to send to the client.
func (p *Peer) NewRequest(info string) (*Request, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s/update?info=%s", p.network, p.addr, info), nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Request{
		request: req.WithContext(ctx),
		client:  p.client,
		cancel:  cancel,
	}, nil
}

// Request encapsulates a way to do the request on the chosen peer.
type Request struct {
	request *http.Request
	client  *http.Client
	cancel  context.CancelFunc
}

// Do the peer request and wait for the response.
func (r *Request) Do() (*http.Response, error) {
	return r.client.Do(r.request)
}

// Cancel allows the cancelling of the peer request.
func (r *Request) Cancel() {
	r.cancel()
}

// URL returns the request URL, which is useful for debugging.
func (r *Request) URL() *url.URL {
	return r.request.URL
}
