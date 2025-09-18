package clacks

import (
	"context"
	"net/http"
)

// GetMessagesFns are used by the client and the server locate the values to be set as the outgoing X-Clacks_Overhead header.
type GetMessagesFn func(context.Context, *http.Request) ([]string, error)

// GetDefaultMessage is the default GetMessagesFn. Returns the value "GNU Terry Pratchett".
func GetDefaultMessage(context.Context, *http.Request) ([]string, error) {
	return []string{DefaultMessage}, nil
}

// HandleSendOnMessagesFn is a type of handler for incoming Clacks Overhead messages with the 'G' code indicating that the message should be sent on.
// Can be used to store messages outside of the current request context so they can be sent on elsewhere.
type HandleSendOnMessagesFn func(context.Context, *http.Request, []string) error
