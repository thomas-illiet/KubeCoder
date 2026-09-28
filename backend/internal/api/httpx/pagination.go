package httpx

import (
	"net/http"
	"strconv"
)

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

// Pagination contains bounded offset pagination parameters parsed from a request.
type Pagination struct {
	Limit  int
	Offset int
}

// Page is the common paginated API response envelope.
type Page[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// ParsePagination reads and bounds limit and offset query parameters.
func ParsePagination(r *http.Request) Pagination {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}
	if offset < 0 {
		offset = 0
	}
	return Pagination{Limit: limit, Offset: offset}
}

// NewPage builds a stable page envelope and normalizes nil items to an empty JSON array.
func NewPage[T any](items []T, total int64, pagination Pagination) Page[T] {
	if items == nil {
		items = make([]T, 0)
	}
	return Page[T]{Items: items, Total: total, Limit: pagination.Limit, Offset: pagination.Offset}
}
