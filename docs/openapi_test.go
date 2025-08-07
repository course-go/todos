package http_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestAPIValidateSchema(t *testing.T) {
	t.Skip() // TODO: kin-openapi does not currently support OpenAPI v3.1
	t.Parallel()

	ctx := t.Context()

	doc, err := openapi3.NewLoader().LoadFromFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("failed loading openapi spec from file: %v", err)
	}

	err = doc.Validate(ctx)
	if err != nil {
		t.Fatalf("failed validation openapi spec: %v", err)
	}
}
