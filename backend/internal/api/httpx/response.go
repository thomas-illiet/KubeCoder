// Package httpx provides transport-level HTTP helpers shared by API domains.
package httpx

import (
	"encoding/json"
	"net/http"
)

// Problem is the common RFC 9457-compatible error response.
type Problem struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Status        int    `json:"status"`
	Detail        string `json:"detail"`
	Instance      string `json:"instance"`
	CorrelationID string `json:"correlation_id"`
}

// WriteProblem writes an RFC 9457-compatible problem response.
func WriteProblem(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{
		Type: "about:blank", Title: title, Status: status, Detail: detail,
		Instance: r.URL.Path, CorrelationID: RequestID(r.Context()),
	})
}

// WriteJSON serializes a value as a JSON HTTP response.
func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
