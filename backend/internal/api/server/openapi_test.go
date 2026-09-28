package server

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestOpenAPIIsValid validates the embedded OpenAPI document.
func TestOpenAPIIsValid(t *testing.T) {
	t.Parallel()
	data, err := openAPIFS.ReadFile("openapi/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	document, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		t.Fatalf("load OpenAPI: %v", err)
	}
	if err := document.Validate(context.Background()); err != nil {
		t.Fatalf("validate OpenAPI: %v", err)
	}

	declaredTags := make(map[string]struct{}, len(document.Tags))
	for _, tag := range document.Tags {
		declaredTags[tag.Name] = struct{}{}
	}
	for path, item := range document.Paths.Map() {
		for method, operation := range item.Operations() {
			if len(operation.Tags) != 1 {
				t.Errorf("%s %s has %d tags, want 1", method, path, len(operation.Tags))
				continue
			}
			if _, ok := declaredTags[operation.Tags[0]]; !ok {
				t.Errorf("%s %s uses undeclared tag %q", method, path, operation.Tags[0])
			}
		}
	}
}
