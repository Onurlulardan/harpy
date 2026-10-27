package migrationsupersession

import (
	"encoding/json"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func decodeCatalogDocument(t *testing.T, encoded []byte, target *catalogDocument) {
	t.Helper()
	if err := json.Unmarshal(encoded, target); err != nil {
		t.Fatal(err)
	}
}

func marshalCatalogDocument(t *testing.T, document catalogDocument) []byte {
	t.Helper()
	encoded, err := attestation.MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	return append(encoded, '\n')
}
