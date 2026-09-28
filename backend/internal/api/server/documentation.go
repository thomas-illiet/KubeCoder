package server

import (
	"embed"
	"net/http"

	"github.com/swaggest/swgui/v5emb"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/httpx"
)

//go:embed openapi/openapi.yaml
var openAPIFS embed.FS

// registerDocumentationRoutes registers the OpenAPI document and Swagger UI.
func registerDocumentationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.yaml", openapi)
	mux.Handle("/docs/", v5emb.New("KubeCoder API", "/openapi.yaml", "/docs/"))
}

// openapi serves the embedded OpenAPI source document.
func openapi(w http.ResponseWriter, r *http.Request) {
	data, err := openAPIFS.ReadFile("openapi/openapi.yaml")
	if err != nil {
		httpx.WriteProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "The API specification is unavailable.")
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
