package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestCatalogOpaqueValuesAreRedacted(t *testing.T) {
	hash, _ := NewCatalogHash(validCatalogHash)
	reference, _ := NewCatalogReference(
		"private_owner", catalogResourceID, catalogVersionID, hash,
	)
	for _, value := range []any{hash, reference} {
		stringer := value.(fmt.Stringer)
		goStringer := value.(fmt.GoStringer)
		for _, formatted := range []string{
			stringer.String(), goStringer.GoString(), fmt.Sprint(value),
			fmt.Sprintf("%s", value), fmt.Sprintf("%q", value),
			fmt.Sprintf("%#v", value), fmt.Sprintf("%x", value),
		} {
			if formatted != queryValueRedaction {
				t.Fatalf("%T formatting leaked catalog data: %q", value, formatted)
			}
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		} else if strings.Contains(err.Error(), "private_owner") ||
			strings.Contains(err.Error(), validCatalogHash) ||
			strings.Contains(err.Error(), catalogResourceID) {
			t.Fatalf("%T marshal error leaked catalog data: %v", value, err)
		}
	}
}

func TestCatalogConstructionErrorsDoNotLeakPrivateValues(t *testing.T) {
	privateOwner := "PrivateOwner"
	privateHash := strings.Repeat("f", 63) + "X"
	if _, err := NewCatalogHash(privateHash); err == nil ||
		strings.Contains(err.Error(), privateHash) {
		t.Fatalf("catalog hash error leaked input: %v", err)
	}
	hash, _ := NewCatalogHash(validCatalogHash)
	if _, err := NewCatalogReference(
		privateOwner, catalogResourceID, catalogVersionID, hash,
	); err == nil || strings.Contains(err.Error(), privateOwner) ||
		strings.Contains(err.Error(), catalogResourceID) {
		t.Fatalf("catalog reference error leaked input: %v", err)
	}
}
