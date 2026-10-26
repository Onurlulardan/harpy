package cutoverv1

import (
	"encoding/json"
	"os"
	"testing"
)

func TestChangeBatchSchemaUsesAuthoritativeByteLimit(t *testing.T) {
	if MaximumChangeBatchBytes != 8*1024*1024 {
		t.Fatalf("change batch byte limit = %d", MaximumChangeBatchBytes)
	}
	document, err := os.ReadFile("change-batch.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(document, &schema); err != nil {
		t.Fatal(err)
	}
	value, ok := schema["x-harpy-maximum-canonical-bytes"].(float64)
	if !ok || int(value) != MaximumChangeBatchBytes {
		t.Fatalf("change batch schema byte limit = %v", schema["x-harpy-maximum-canonical-bytes"])
	}
}
