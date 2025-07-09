package clacks

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

// Identical to http.DefaultClient except it will send "GNU Terry Pratchett"
// as a X-Clacks-Overhead header with every request.
func DefaultClient() *http.Client {
	return &http.Client{
		Transport: &RoundTripper{
			GetOverheadMessages: GetDefaultMessage,
			Transport:           http.DefaultTransport,
		},
	}
}

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

type GetMessagesFn func(context.Context, *http.Request) ([]string, error)

func GetDefaultMessage(context.Context, *http.Request) ([]string, error) {
	return []string{DefaultMessage}, nil
}
