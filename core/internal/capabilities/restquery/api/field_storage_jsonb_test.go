package api

import (
	"encoding/json"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const jsonbStorageKey = "de0eddb4-a8e8-4df9-9499-0a581c02f500"

func TestJSONBFieldStorageSealsCanonicalStableKey(t *testing.T) {
	document, err := sqlidentifier.MintStatic("profile_values")
	if err != nil {
		t.Fatal(err)
	}
	storage, err := NewJSONBFieldStorage(document, jsonbStorageKey)
	if err != nil {
		t.Fatalf("new JSONB storage: %v", err)
	}
	key, present, err := storage.JSONBKey()
	if err != nil || !present || key.String() != jsonbStorageKey {
		t.Fatalf("JSONB key = %v, %v, %v", key, present, err)
	}
	column, err := storage.Column()
	if err != nil || column != document {
		t.Fatal("JSONB document column was not preserved")
	}
	if _, err := json.Marshal(storage); err == nil {
		t.Fatal("JSONB storage exposed its sealed path")
	}
}

func TestJSONBFieldStorageRejectsRequestSelectedPathShapes(t *testing.T) {
	document, _ := sqlidentifier.MintStatic("profile_values")
	for _, key := range []string{
		"", uuid.Nil.String(), "DE0EDDB4-A8E8-4DF9-9499-0A581C02F500",
		jsonbStorageKey + "' OR TRUE --", "nickname",
	} {
		if _, err := NewJSONBFieldStorage(document, key); err == nil {
			t.Fatalf("unsafe JSONB key %q was accepted", key)
		}
	}
}

func TestJSONBFieldStorageRejectsCompanionTamper(t *testing.T) {
	document, _ := sqlidentifier.MintStatic("profile_values")
	key := uuid.MustParse(jsonbStorageKey)
	storage := FieldStorage{
		column: document, jsonbKey: key, hasJSONB: true,
		folded: document, hasFolded: true, validated: true,
	}
	if storage.Validate() == nil {
		t.Fatal("JSONB storage accepted a text companion")
	}
}
