package clacks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Middleware will include "GNU Terry Pratchett" as an overhead in every response by default.
// More complex behaviour can be enabled by passing options.
func Middleware(optFns ...MiddlewareOptFn) func(http.Handler) http.Handler {
	opts := &MiddlewareOpts{
		GetOverheadMessages: GetDefaultMessage,
	}

	for _, optFn := range optFns {
		optFn(opts)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			msgs, err := opts.GetOverheadMessages(r.Context(), r)
			if err != nil {
				// If we can't get the outgoing clacks overhead values log the error but handle the request as normal.
				// No point in killing the request for the sake of this header - Sorry Terry!
				if opts.Logger != nil {
					opts.Logger.Error("Clacks Overhead Middleware: could not get overhead messages", err)
				}

				next.ServeHTTP(w, r)
				return
			}

			in := r.Header.Values(OverheadHeaderKey)
			if len(in) == 0 {
				for _, msg := range msgs {
					w.Header().Add(OverheadHeaderKey, msg)
				}

				next.ServeHTTP(w, r)
				return
			}

			// Process incoming clacks overhead messages in request headers.
			out := make(map[string]struct{}, len(in)+len(msgs))
			var sendOnMsgs []string
			for _, msg := range in {
				codes := GetCodesFromOverhead(msg)

				// Log if allowed.
				if opts.Logger != nil && !strings.Contains(codes, CodeNoLogs) {
					opts.Logger.Print(fmt.Sprintf("Received Clacks Overhead message: %q", msg))
				}

				// Turn around
				if strings.Contains(codes, CodeTurnAround) {
					out[msg] = struct{}{}
				}

				// Send on.
				if opts.SendOnHandler != nil && strings.Contains(codes, CodeSendOn) {
					sendOnMsgs = append(sendOnMsgs, msg)
				}
			}

			// The intent here is to enable clacks overhead messages with the 'G' code (send on) to "escape" the current request
			// context and perhaps be included in other requests/responses. Always going home.
			if opts.SendOnHandler != nil {
				err = opts.SendOnHandler(r.Context(), r, sendOnMsgs)
				if err != nil && opts.Logger != nil {
					opts.Logger.Error("Clacks Overhead Middleware: could not handle send-on messages", sendOnMsgs, err)
				}
			}

			// Process outgoing clacks overhead messages.
			// Map to prevent duplication.
			for _, msg := range msgs {
				out[msg] = struct{}{}
			}

			for msg := range out {
				w.Header().Add(OverheadHeaderKey, msg)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetCodesFromOverhead extracts the clacks codes from an overhead message.
// Passing "GNU incoming msg" returns "GNU".
// The clacks protocol is not well-defined in the Discworld corpus.
// We make some assumptions when extracting codes:
//  1. Clacks codes go at the start of the message and are followed by a space.
//  2. All clacks codes are single uppercase letters.
//  3. We might receive naked overhead messages (i.e. not preceded by codes); we'll do our best to filter these out.
func GetCodesFromOverhead(msg string) string {
	if msg == "" {
		return ""
	}

	ss := strings.Split(msg, " ")
	for _, c := range ss[0] {
		if c < 'A' || 'Z' < c {
			return ""
		}
	}
	return ss[0]
}

// MiddlewareOpts are the options for the clacks overhead middleware.
type MiddlewareOpts struct {
	// GetOverheadMessages will be called to get the overhead messages to include in the response.
	GetOverheadMessages GetMessagesFn
	Logger              Logger
	// SendOnHandler will be called on any X-Clacks-Overhead messages with the 'G' code in the incoming request (unless nil).
	SendOnHandler HandleSendOnMessagesFn
}

type MiddlewareOptFn func(*MiddlewareOpts)

func WithLogger(l Logger) MiddlewareOptFn {
	return func(o *MiddlewareOpts) {
		o.Logger = l
	}
}

func WithOverheadMessages(msgs ...string) MiddlewareOptFn {
	return func(o *MiddlewareOpts) {
		o.GetOverheadMessages = func(ctx context.Context, r *http.Request) ([]string, error) {
			return msgs, nil
		}
	}
}

func WithSendOnHandler(handler HandleSendOnMessagesFn) MiddlewareOptFn {
	return func(o *MiddlewareOpts) {
		o.SendOnHandler = handler
	}
}

type Logger interface {
	Print(v ...any)
	Error(v ...any)
}
