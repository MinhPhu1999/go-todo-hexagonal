package platform

import (
	"encoding/json"
	"testing"
)

func TestOpenAPIJSONIsValid(t *testing.T) {
	var document map[string]any
	if err := json.Unmarshal([]byte(OpenAPIJSON), &document); err != nil {
		t.Fatalf("OpenAPIJSON is invalid JSON: %v", err)
	}

	if document["openapi"] != "3.0.3" {
		t.Fatalf("unexpected OpenAPI version: %v", document["openapi"])
	}
}
