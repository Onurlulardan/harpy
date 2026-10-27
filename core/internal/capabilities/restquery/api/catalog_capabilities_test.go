package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCatalogCapabilitiesExactCountAndZeroSafety(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		capabilities, err := NewCatalogCapabilities(enabled)
		if err != nil || capabilities.Validate() != nil {
			t.Fatalf("valid catalog capabilities rejected: %v", err)
		}
		actual, err := capabilities.ExactCountEnabled()
		if err != nil || actual != enabled {
			t.Fatal("exact count capability did not round trip")
		}
		if fmt.Sprint(capabilities) != queryValueRedaction {
			t.Fatal("catalog capabilities formatting was not redacted")
		}
		if _, err := json.Marshal(capabilities); err == nil {
			t.Fatal("catalog capabilities silently marshaled")
		}
	}
	var zero CatalogCapabilities
	if zero.Validate() == nil {
		t.Fatal("zero catalog capabilities validated")
	}
	if _, err := zero.ExactCountEnabled(); err == nil {
		t.Fatal("zero catalog capabilities accessor succeeded")
	}
}
