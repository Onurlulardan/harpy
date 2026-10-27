package aiinferencev1

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestInferenceSchemasCompileAndRemainClosed(t *testing.T) {
	for _, name := range []string{"request.schema.json", "result.schema.json"} {
		document, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
		if err != nil {
			t.Fatal(err)
		}
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(name, decoded); err != nil {
			t.Fatal(err)
		}
		if _, err := compiler.Compile(name); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !strings.Contains(string(document), "additionalProperties") &&
			!strings.Contains(string(document), "unevaluatedProperties") {
			t.Fatalf("%s does not close provider-controlled objects", name)
		}
	}
}

func TestInferenceOpenAPIPublishesOnlyThreePrivateCoreRoutes(t *testing.T) {
	document, err := os.ReadFile("inference.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract map[string]any
	if err := json.Unmarshal(document, &contract); err != nil {
		t.Fatal(err)
	}
	paths, ok := contract["paths"].(map[string]any)
	want := []string{"/internal/ai/v1/generate", "/internal/ai/v1/generate/stream",
		"/internal/ai/v1/embed"}
	if !ok || len(paths) != len(want) {
		t.Fatalf("inference paths=%#v", paths)
	}
	for _, path := range want {
		if paths[path] == nil {
			t.Fatalf("missing private route %s", path)
		}
	}
	lower := strings.ToLower(string(document))
	for _, required := range []string{"core-ai-inference.v1", "harpy-ai-provider-v1",
		"coreservicecredential", "text/event-stream"} {
		if !strings.Contains(lower, strings.ToLower(required)) {
			t.Fatalf("OpenAPI misses %s", required)
		}
	}
	for _, forbidden := range []string{"/api/ai-chat", "client_bearer", "/public/"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("OpenAPI exposes forbidden surface %s", forbidden)
		}
	}
}
