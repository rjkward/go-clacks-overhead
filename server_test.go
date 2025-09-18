package clacks

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	cases := []struct {
		name          string
		defaultOpts   bool
		out           []string
		outErr        error
		in            []string
		sendOnErr     error
		wantHeaders   []string
		wantLogs      []string
		wantErrorLogs []string
		wantSendOn    []string
	}{
		{
			name:        "should add default header with default settings",
			defaultOpts: true,
			wantHeaders: []string{"GNU Terry Pratchett"},
		},
		{
			name:        "should turn around incoming with default settings",
			defaultOpts: true,
			in:          []string{"U Test Incoming"},
			wantHeaders: []string{"GNU Terry Pratchett", "U Test Incoming"},
		},
		{
			name:        "should turn around incoming and handle dups with default settings",
			defaultOpts: true,
			in:          []string{"U Test Incoming", "GNU Terry Pratchett"},
			wantHeaders: []string{"GNU Terry Pratchett", "U Test Incoming"},
		},
		{
			name: "should add specified headers",
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
			name:        "should add header with incoming msgs and logs",
			out:         []string{"outgoing msg"},
			in:          []string{"incoming msg"},
			wantHeaders: []string{"outgoing msg"},
			wantLogs:    []string{"Received Clacks Overhead message: \"incoming msg\"\n"},
		},
		{
			name: "should respond with 'turned around' incoming headers with U code and not duplicate",
			out: []string{
				"GNU Terry Pratchett",
				"M. John Harrison",
			},
			in: []string{
				"GNU Terry Pratchett",
				"NU Ursula Le Guin",
				"K. A. Applegate",
			},
			wantHeaders: []string{
				"GNU Terry Pratchett",
				"NU Ursula Le Guin",
				"M. John Harrison",
			},
			wantLogs: []string{
				"Received Clacks Overhead message: \"K. A. Applegate\"\n",
			},
			wantSendOn: []string{
				"GNU Terry Pratchett",
			},
		},
		{
			name: "should not log incoming messages with N code",
			out: []string{
				"GNU Terry Pratchett",
			},
			in: []string{
				"N Ursula Le Guin",
				"K. A. Applegate",
				"M. John Harrison",
			},
			wantHeaders: []string{
				"GNU Terry Pratchett",
			},
			wantLogs: []string{
				"Received Clacks Overhead message: \"K. A. Applegate\"\n",
				"Received Clacks Overhead message: \"M. John Harrison\"\n",
			},
		},
		{
			name:          "should log GetOverheadMessages error and continue",
			outErr:        errors.New("test error"),
			wantErrorLogs: []string{"Clacks Overhead Middleware: could not get overhead messages test error\n"},
		},
		{
			name: "should handle send on messages",
			out:  []string{"GNU outgoing msg"},
			in: []string{
				"GNU incoming msg 1",
				"G incoming msg 2",
				"NU incoming msg 3",
				"incoming msg 4",
			},
			wantHeaders: []string{
				"GNU outgoing msg",
				"GNU incoming msg 1",
				"NU incoming msg 3",
			},
			wantLogs: []string{
				"Received Clacks Overhead message: \"G incoming msg 2\"\n",
				"Received Clacks Overhead message: \"incoming msg 4\"\n",
			},
			wantSendOn: []string{
				"GNU incoming msg 1",
				"G incoming msg 2",
			},
		},
		{
			name:      "should log SendOnHandler error and continue",
			out:       []string{"GNU outgoing msg"},
			in:        []string{"GNU incoming msg"},
			sendOnErr: errors.New("test send on error"),
			wantHeaders: []string{
				"GNU outgoing msg",
				"GNU incoming msg",
			},
			wantErrorLogs: []string{"Clacks Overhead Middleware: could not handle send-on messages [GNU incoming msg] test send on error\n"},
		},
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			logger := &mockLogger{}

			var gotSendOn []string
			var middleware func(http.Handler) http.Handler
			if c.defaultOpts {
				middleware = Middleware()
			} else {
				middleware = Middleware(func(mo *MiddlewareOpts) {
					mo.GetOverheadMessages = func(ctx context.Context, r *http.Request) ([]string, error) {
						return c.out, c.outErr
					}
					mo.Logger = logger
					mo.SendOnHandler = func(ctx context.Context, r *http.Request, msgs []string) error {
						gotSendOn = append(gotSendOn, msgs...)
						return c.sendOnErr
					}
				})
			}

			wrappedHandler := middleware(inner)

			req := httptest.NewRequest("GET", "/", http.NoBody)
			for _, i := range c.in {
				req.Header.Add(OverheadHeaderKey, i)
			}

			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)

			got := w.Header().Values(OverheadHeaderKey)
			require.ElementsMatch(t, c.wantHeaders, got)

			if c.sendOnErr == nil {
				require.ElementsMatch(t, c.wantSendOn, gotSendOn)
			}

			require.ElementsMatch(t, c.wantLogs, logger.Logs)
			require.ElementsMatch(t, c.wantErrorLogs, logger.Errors)
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

func TestWithLogger(t *testing.T) {
	logger := &mockLogger{}
	opts := &MiddlewareOpts{}
	WithLogger(logger)(opts)
	require.Equal(t, logger, opts.Logger)
}

func TestWithOverheadMessages(t *testing.T) {
	msgs := []string{"test1", "test2"}
	opts := &MiddlewareOpts{}
	WithOverheadMessages(msgs...)(opts)
	got, err := opts.GetOverheadMessages(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, msgs, got)
}

func TestWithSendOnHandler(t *testing.T) {
	var out []string
	handler := func(ctx context.Context, r *http.Request, msgs []string) error {
		out = append(out, msgs...)
		return nil
	}
	opts := &MiddlewareOpts{}
	WithSendOnHandler(handler)(opts)
	opts.SendOnHandler(context.Background(), nil, []string{"test"})
	require.Equal(t, []string{"test"}, out)
}

type mockLogger struct {
	Logs   []string
	Errors []string
}

func (m *mockLogger) Print(v ...any) {
	m.Logs = append(m.Logs, fmt.Sprintln(v...))
}

func (m *mockLogger) Error(v ...any) {
	m.Errors = append(m.Errors, fmt.Sprintln(v...))
}
