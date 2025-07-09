package clacks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	cases := []struct {
		name        string
		optFn       MiddlewareOptFn
		incoming    []string
		wantHeaders []string
	}{
		{
			name:        "should add default header with default settings",
			optFn:       func(opts *MiddlewareOpts) {},
			wantHeaders: []string{"GNU Terry Pratchett"},
		},
		{
			name: "should add specified headers with default settings",
			optFn: func(opts *MiddlewareOpts) {
				opts.GetOverheadMessages = func(ctx context.Context, r *http.Request) ([]string, error) {
					return []string{
						"test",
						"test2",
						"test3",
					}, nil
				}
			},
			wantHeaders: []string{
				"test",
				"test2",
				"test3",
			},
		},
		{
			name:        "should add default header with incoming msgs",
			optFn:       func(opts *MiddlewareOpts) {},
			incoming:    []string{"test"},
			wantHeaders: []string{"GNU Terry Pratchett"},
		},
		{
			name: "should respond with 'turned around' incoming headers with U code and not duplicate",
			optFn: func(opts *MiddlewareOpts) {
				opts.GetOverheadMessages = func(ctx context.Context, r *http.Request) ([]string, error) {
					return []string{"GNU Carmen Sandiego"}, nil
				}
			},
			incoming: []string{
				"GNU Terry Pratchett",
				"U Joe Bloggs",
				"GNU Carmen Sandiego",
				"Jane Doe",
			},
			wantHeaders: []string{
				"GNU Terry Pratchett",
				"U Joe Bloggs",
				"GNU Carmen Sandiego",
			},
		},
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			middlware := Middleware(c.optFn)
			wrappedHandler := middlware(inner)

			req := httptest.NewRequest("GET", "/", nil)
			for _, i := range c.incoming {
				req.Header.Add(OverheadHeaderKey, i)
			}

			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			got := w.Header().Values(OverheadHeaderKey)
			require.ElementsMatch(t, c.wantHeaders, got)
			require.Equal(t, http.StatusOK, w.Code)
		})
	}
}
