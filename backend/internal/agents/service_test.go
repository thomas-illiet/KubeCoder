package agents

import (
	"errors"
	"testing"
)

// TestValidatedAgent covers normalization, capability de-duplication, and required fields.
func TestValidatedAgent(t *testing.T) {
	t.Parallel()
	input := Input{Name: " Atlas ", Description: " General agent ", RuntimeAdapter: " opencode ", RuntimeVersion: "1.2", Image: "example/agent@sha256:abc", Provider: "openai", Model: "gpt", SystemPrompt: "Help", Capabilities: []string{"tests", " tests ", ""}, CPUMillis: 1000, MemoryMB: 2048, StorageMB: 4096, MaxDurationSeconds: 60, Active: true}
	item, err := validated(input)
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "Atlas" || len(item.Capabilities) != 1 || item.Capabilities[0] != "tests" {
		t.Fatalf("unexpected normalized agent: %#v", item)
	}
	input.CPUMillis = 0
	if _, err := validated(input); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
