package postgres

import (
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestExtensionPlacementFollowsDiscoveredOrigin(t *testing.T) {
	t.Parallel()
	accepted := []struct {
		origin runtimev1.DatabaseOrigin
		schema string
		owner  string
	}{
		{runtimev1.DatabaseOriginFresh, "platform_ext", "harpy_extension_admin"},
		{runtimev1.DatabaseOriginV1Adopted, "public", "legacy_owner"},
		{runtimev1.DatabaseOriginV1Adopted, "public", "harpy_legacy_extension_owner"},
	}
	for _, test := range accepted {
		if err := validateExtensionPlacement(test.origin, test.schema, test.owner); err != nil {
			t.Fatalf("valid placement rejected: %+v: %v", test, err)
		}
	}
	if err := validateExtensionPlacement(
		runtimev1.DatabaseOriginV1Adopted, "public", "harpy_runtime",
	); err == nil {
		t.Fatal("runtime role was accepted as adopted extension owner")
	}
}
