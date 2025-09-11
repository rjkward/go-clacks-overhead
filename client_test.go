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

func TestRoundTripper_Copy(t *testing.T) {
	mt := new(mockTransport)
	rt := &RoundTripper{
		GetOverheadMessages: func(ctx context.Context, r *http.Request) ([]string, error) {
			return []string{"test"}, nil
		},
		Transport: mt,
	}

	req1 := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	res1, err := rt.RoundTrip(req1)
	require.NoError(t, err)
	require.True(t, mt.Called)
	require.Equal(t, http.StatusOK, res1.StatusCode)
	require.NotSame(t, req1, mt.Req)
	require.Equal(t, []string{"test"}, mt.Req.Header.Values(OverheadHeaderKey))

	// reset for non copy test
	mt = new(mockTransport)
	rt = &RoundTripper{
		GetOverheadMessages: func(ctx context.Context, r *http.Request) ([]string, error) {
			return []string{"test"}, nil
		},
		Transport: mt,
		// disable copying of request
		UseOriginalRequest: true,
	}

	req2 := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	res2, err := rt.RoundTrip(req2)
	require.NoError(t, err)
	require.True(t, mt.Called)
	require.Equal(t, http.StatusOK, res2.StatusCode)
	require.Same(t, req2, mt.Req)
	require.Equal(t, []string{"test"}, mt.Req.Header.Values(OverheadHeaderKey))
}

type mockTransport struct {
	Req    *http.Request
	Called bool
}

func (m *mockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	m.Called = true
	m.Req = r
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
	}, nil
}
