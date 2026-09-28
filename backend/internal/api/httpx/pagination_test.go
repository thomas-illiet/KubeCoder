package httpx

import (
	"net/http/httptest"
	"testing"
)

// TestParsePagination verifies defaults, explicit values, bounds, and invalid input handling.
func TestParsePagination(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		query      string
		pagination Pagination
	}{
		{name: "defaults", pagination: Pagination{Limit: DefaultPageLimit}},
		{name: "explicit", query: "?limit=25&offset=50", pagination: Pagination{Limit: 25, Offset: 50}},
		{name: "bounded", query: "?limit=1000&offset=-10", pagination: Pagination{Limit: MaxPageLimit}},
		{name: "invalid", query: "?limit=invalid&offset=invalid", pagination: Pagination{Limit: DefaultPageLimit}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest("GET", "/resources"+test.query, nil)
			if got := ParsePagination(request); got != test.pagination {
				t.Fatalf("pagination = %#v, want %#v", got, test.pagination)
			}
		})
	}
}

// TestNewPageNormalizesNilItems verifies that paginated JSON uses an empty array instead of null.
func TestNewPageNormalizesNilItems(t *testing.T) {
	t.Parallel()
	page := NewPage[string](nil, 3, Pagination{Limit: 20, Offset: 0})
	if page.Items == nil {
		t.Fatal("items must be an empty slice")
	}
	if page.Total != 3 || page.Limit != 20 || page.Offset != 0 {
		t.Fatalf("page = %#v", page)
	}
}
