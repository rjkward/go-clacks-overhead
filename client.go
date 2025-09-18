package clacks

import (
	"net/http"

	"github.com/pkg/errors"
)

// DefaultClient is identical to http.DefaultClient except it will send "GNU Terry Pratchett"
// as a X-Clacks-Overhead header with every request.
var DefaultClient = &http.Client{
	Transport: DefaultTransport,
}

// DefaultTransport is identical to http.DefaultTransport except it will send "GNU Terry Pratchett"
// as a X-Clacks-Overhead header with every request.
var DefaultTransport = &RoundTripper{
	GetOverheadMessages: GetDefaultMessage,
	Transport:           http.DefaultTransport,
}

// RoundTripper is an http.RoundTripper that will add X-Clacks-Overhead headers to requests.
// By default it will add "GNU Terry Pratchett" to every request, but this can be changed by
// setting GetOverheadMessages. Unless UseOriginalRequest is set to true, the request will be cloned.
type RoundTripper struct {
	GetOverheadMessages GetMessagesFn
	Transport           http.RoundTripper
	UseOriginalRequest  bool
}

func (c *RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var msgs []string
	if c.GetOverheadMessages != nil {
		var err error
		msgs, err = c.GetOverheadMessages(r.Context(), r)
		if err != nil {
			return nil, errors.Wrap(err, "could not get clacks overhead messages for request")
		}
	}

	if !c.UseOriginalRequest {
		r = r.Clone(r.Context())
	}

	for _, msg := range msgs {
		r.Header.Add(OverheadHeaderKey, msg)
	}

	return c.Transport.RoundTrip(r)
}
