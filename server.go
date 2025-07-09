package clacks

import (
	"fmt"
	"net/http"
	"strings"
)

// Middleware implements the Clacks Overhead protocol. It will handle incoming overhead messages
// appropriately and include "GNU Terry Pratchett" as an overhead in every response by default.
func Middleware(optFns ...MiddlewareOptFn) func(http.Handler) http.Handler {
	opts := &MiddlewareOpts{
		OverheadMessages: []string{DefaultMessage},
	}

	for _, optFn := range optFns {
		optFn(opts)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Process incoming clacks overhead messages.
			in := r.Header.Values(OverheadHeaderKey)
			out := make(map[string]struct{}, len(in)+len(opts.OverheadMessages))
			for _, msg := range in {
				codes := GetCodesFromOverhead(msg)

				// Log if allowed.
				if opts.Logger != nil && !strings.Contains(codes, CodeNoLogs) {
					opts.Logger.Print(fmt.Sprintf("Received Clacks Overhead message: %q", msg))
				}

				// Send on.
				if strings.Contains(codes, CodeSendOn) || strings.Contains(codes, CodeTurnAround) {
					out[msg] = struct{}{}
				}
			}

			// Process outgoing clacks overhead messages.
			for _, msg := range opts.OverheadMessages {
				out[msg] = struct{}{}
			}

			for msg := range out {
				w.Header().Add(OverheadHeaderKey, msg)
			}

			next.ServeHTTP(w, r)
		})
	}
}

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
		if c < 'A' || c > 'Z' {
			return ""
		}
	}
	return ss[0]
}

type MiddlewareOpts struct {
	OverheadMessages []string
	Logger           Logger
}

type MiddlewareOptFn func(*MiddlewareOpts)

func WithLogger(l Logger) MiddlewareOptFn {
	return func(o *MiddlewareOpts) {
		o.Logger = l
	}
}

func WithOverheadMessages(msgs ...string) MiddlewareOptFn {
	return func(o *MiddlewareOpts) {
		o.OverheadMessages = msgs
	}
}

type Logger interface {
	Print(v ...any)
}
