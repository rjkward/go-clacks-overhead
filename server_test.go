package clacks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := Middleware()(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	got := w.Header().Get(OverheadHeaderKey)
	want := DefaultMessage
	require.Equal(t, want, got)
	require.Equal(t, http.StatusOK, w.Code)
}
