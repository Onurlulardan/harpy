package application

import (
	"errors"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func TestReferenceEvidenceRequiresExactLayoutOwnerManifest(t *testing.T) {
	target := ReferenceTarget{
		Resource: layoutapi.ResourceUser,
		FieldID:  uuid.MustParse("33333333-3333-4333-8333-333333333333"),
	}
	valid := ReferenceEvidence{
		Target: target, ManifestSHA256: [32]byte{1}, CheckedOwners: 1,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid reference evidence rejected: %v", err)
	}
	invalid := []ReferenceEvidence{
		{},
		{Target: target, CheckedOwners: 1},
		{Target: target, ManifestSHA256: [32]byte{1}, CheckedOwners: 2},
		{Target: target, ManifestSHA256: [32]byte{1}, CheckedOwners: 1,
			LiveReferences: -1},
	}
	for _, evidence := range invalid {
		if !errors.Is(evidence.Validate(), layoutapi.ErrUnavailable) {
			t.Fatalf("invalid reference evidence accepted: %+v", evidence)
		}
	}
}
