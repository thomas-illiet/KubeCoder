package repositories

import (
	"errors"
	"testing"
)

// TestNormalizeCloneURL verifies supported schemes and stable URL normalization.
func TestNormalizeCloneURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, input, want string
		invalid           bool
	}{
		{name: "https", input: "https://GitHub.com/Org/Repo.git/", want: "https://github.com/Org/Repo"},
		{name: "ssh", input: "ssh://git@gitlab.example/Org/Repo.git", want: "ssh://git@gitlab.example/Org/Repo"},
		{name: "credentials", input: "https://token@github.com/org/repo", invalid: true},
		{name: "scp syntax", input: "git@github.com:org/repo.git", invalid: true},
		{name: "unsupported", input: "http://github.com/org/repo", invalid: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeCloneURL(test.input)
			if test.invalid {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("expected ErrInvalid, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

// TestValidProvider locks the public provider allowlist.
func TestValidProvider(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"github", "gitlab", "bitbucket"} {
		if !validProvider(value) {
			t.Errorf("expected %q to be valid", value)
		}
	}
	if validProvider("azure") {
		t.Error("unexpected Azure provider support")
	}
}
