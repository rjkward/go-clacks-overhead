package clacks

import (
	"context"
	"net/http"
)

// Used by the client and the server to set the value of the X-Clacks_Overhead header.
type GetMessagesFn func(context.Context, *http.Request) ([]string, error)

// Default GetMessagesFn. Returns the value "GNU Terry Pratchett".
func GetDefaultMessage(context.Context, *http.Request) ([]string, error) {
	return []string{DefaultMessage}, nil
}

// Handler for incoming Clacks Overhead messages with the 'G' code indicating that the message should be sent on.
// Can be used to store messages outside of the current request context so they can be sent on elsewhere.
type HandleSendOnMessagesFn func(context.Context, *http.Request, []string) error
