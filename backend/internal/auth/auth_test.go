package auth

import "testing"

// TestBearerToken verifies strict Authorization header parsing.
func TestBearerToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		header string
		want   string
		ok     bool
	}{
		{name: "valid", header: "Bearer token-value", want: "token-value", ok: true},
		{name: "case insensitive", header: "bearer token-value", want: "token-value", ok: true},
		{name: "missing", header: "", ok: false},
		{name: "basic", header: "Basic token-value", ok: false},
		{name: "extra", header: "Bearer one two", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BearerToken(tt.header)
			if (err == nil) != tt.ok || got != tt.want {
				t.Fatalf("BearerToken() = %q, %v", got, err)
			}
		})
	}
}
