package clacks

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRountTripper(t *testing.T) {
	cases := []struct {
		name  string
		msgs  []string
		err   error
		nilFn bool
	}{
		{
			name:  "should succeed with nil func",
			nilFn: true,
		},
		{
			name: "should succeed with nil slice",
			msgs: nil,
		},
		{
			name: "should succeed with empty slice",
			msgs: []string{},
		},
		{
			name: "should succeed with 1 message",
			msgs: []string{"test message"},
		},
		{
			name: "should succeed with multiple messages",
			msgs: []string{"test message",
				"another message",
				"yet another message",
			},
		},
		{
			name: "should fail with internal error",
			err:  errors.New("test error"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got := r.Header.Values(OverheadHeaderKey)
				require.ElementsMatch(t, c.msgs, got)
			}))
			defer server.Close()

			var fn GetMessagesFn
			if !c.nilFn {
				fn = func(ctx context.Context, r *http.Request) ([]string, error) {
					return c.msgs, c.err
				}
			}
			rt := &RoundTripper{
				GetOverheadMessages: fn,
				Transport:           http.DefaultTransport,
			}

			req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
			require.NoError(t, err)
			res, err := rt.RoundTrip(req)
			if c.err != nil {
				require.ErrorIs(t, err, c.err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, res.StatusCode)
			}
		})
	}
}
