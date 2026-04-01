//go:build webembed

package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

type Handler struct {
	fileServer http.Handler
	indexHTML  []byte
}

func NewHandler() (*Handler, error) {
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}

	indexHTML, err := distFS.ReadFile("dist/index.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		fileServer: http.FileServer(http.FS(subFS)),
		indexHTML:  indexHTML,
	}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/assets/") {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(h.indexHTML)
}
