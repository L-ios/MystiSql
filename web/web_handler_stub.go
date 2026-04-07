//go:build !webembed

package webui

import (
	"net/http"
)

// Handler is a no-op when WebUI is not embedded (built without -tags webembed).
type Handler struct{}

// NewHandler returns nil when WebUI assets are not embedded.
// server.go already handles nil webuiHandler gracefully.
func NewHandler() (*Handler, error) {
	return nil, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("WebUI not available (built without -tags webembed)\n"))
}
