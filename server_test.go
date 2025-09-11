package clacks

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	cases := []struct {
		name        string
		defaultOpts bool
		out         []string
		in          []string
		wantHeaders []string
		wantSendOn  []string
		wantLogs    []string
	}{
		{
			name:        "should add default header with default settings",
			defaultOpts: true,
			wantHeaders: []string{"GNU Terry Pratchett"},
		},
		{
			name: "should add specified headers with default settings",
			out: []string{
				"test",
				"test2",
				"test3",
			},
			wantHeaders: []string{
				"test",
				"test2",
				"test3",
			},
		},
		{
			name:        "should add default header with incoming msgs",
			in:          []string{"test"},
			wantHeaders: []string{"GNU Terry Pratchett"},
		},
		{
			name: "should respond with 'turned around' incoming headers with U code and not duplicate",
			out: []string{
				"GNU Terry Pratchett",
				"GNU M. John Harrison",
			},
			in: []string{
				"GNU Terry Pratchett",
				"U Ursula Le Guin",
				"K. A. Applegate",
			},
			wantHeaders: []string{
				"GNU Terry Pratchett",
				"U Ursula Le Guin",
				"GNU M. John Harrison",
			},
		},
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			logger := &mockLogger{}

			var middleware func(http.Handler) http.Handler
			if c.defaultOpts {
				middleware = Middleware()
			} else {
				middleware = Middleware(func(mo *MiddlewareOpts) {
					mo.GetOverheadMessages = func(ctx context.Context, r *http.Request) ([]string, error) {
						return c.out, nil
					}
					mo.Logger = logger
				})
			}

			wrappedHandler := middleware(inner)

			req := httptest.NewRequest("GET", "/", nil)
			for _, i := range c.in {
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

func TestGetCodesFromOverhead(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "should return empty string when empty",
			in:   "",
			want: "",
		},
		{
			name: "should return empty string when first word contains non-code characters",
			in:   "Terry Pratchett",
			want: "",
		},
		{
			name: "should return single code",
			in:   "A Terry Pratchett",
			want: "A",
		},
		{
			name: "should return multiple codes",
			in:   "ABC Terry Pratchett",
			want: "ABC",
		},
		{
			name: "should ignore codes after spaces",
			in:   " ABC Terry Pratchett",
			want: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := GetCodesFromOverhead(c.in)
			require.Equal(t, c.want, got)
		})
	}
}

type mockLogger struct {
	Logs   []string
	Errors []string
}

func (m *mockLogger) Print(v ...any) {
	m.Logs = append(m.Logs, fmt.Sprint(v...))
}

func (m *mockLogger) Error(v ...any) {
	m.Errors = append(m.Errors, fmt.Sprint(v...))
}
