package postgres

import (
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestOriginIsDiscoveredFromContractualVectorPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		schema string
		origin runtimev1.DatabaseOrigin
	}{{"platform_ext", runtimev1.DatabaseOriginFresh}, {"public", runtimev1.DatabaseOriginV1Adopted}}
	for _, test := range tests {
		origin, err := originForVectorSchema(test.schema)
		if err != nil || origin != test.origin {
			t.Fatalf("schema %q origin = %q, %v", test.schema, origin, err)
		}
	}
	if _, err := originForVectorSchema("caller_supplied"); err == nil {
		t.Fatal("untrusted vector placement was accepted")
	}
}
