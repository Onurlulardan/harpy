package postgres

import (
	"slices"
	"testing"
)

func TestDecodeExtensionCapabilitiesUsesDatabaseJSONProjection(t *testing.T) {
	want := []string{"trigram_search", "vector_similarity"}
	got, err := decodeExtensionCapabilities(
		`["trigram_search","vector_similarity"]`,
	)
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("extension capabilities = %#v err=%v", got, err)
	}
	for _, invalid := range []string{`{"trigram_search":true}`, `null`, `{`} {
		if _, err := decodeExtensionCapabilities(invalid); err == nil {
			t.Fatalf("invalid extension capabilities accepted: %s", invalid)
		}
	}
}
