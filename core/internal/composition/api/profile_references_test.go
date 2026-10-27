package api

import (
	"context"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	layoutmodule "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type profileReferenceFixture struct {
	target   layoutmodule.ReferenceTarget
	evidence layoutmodule.ReferenceEvidence
}

func (fixture *profileReferenceFixture) FieldReferenceEvidence(
	_ context.Context,
	_ platformtx.TenantScope,
	target layoutmodule.ReferenceTarget,
) (layoutmodule.ReferenceEvidence, error) {
	fixture.target = target
	return fixture.evidence, nil
}

func TestProfileReferencesMapLayoutEvidence(t *testing.T) {
	fieldID := [16]byte{1, 2, 3, 4, 5, 6, 0x40, 8, 0x80, 10, 11, 12, 13, 14, 15, 16}
	digest := [32]byte{1}
	fixture := &profileReferenceFixture{evidence: layoutmodule.ReferenceEvidence{
		Target: layoutmodule.ReferenceTarget{
			Resource: layoutmodule.ResourceUser, FieldID: fieldID,
		},
		ManifestSHA256: digest, CheckedOwners: 1, LiveReferences: 0,
	}}
	references := newProfileReferences(fixture)
	target := iammodule.ProfileRetirementTarget{
		Resource: iammodule.ProfileResourceUser, FieldID: fieldID,
		ExpectedFieldRevision: 3,
	}
	evidence, err := references.AssertRetirable(
		context.Background(), platformtx.TenantScope{}, target,
	)
	if err != nil {
		t.Fatal(err)
	}
	if fixture.target.Resource != layoutmodule.ResourceUser ||
		fixture.target.FieldID != fieldID || evidence.Target != target ||
		evidence.ManifestSHA256 != digest || evidence.CheckedOwners != 1 ||
		evidence.LiveReferences != 0 {
		t.Fatalf("mapped evidence = %+v, layout target = %+v", evidence, fixture.target)
	}
}

func TestProfileReferencesRejectUnknownResource(t *testing.T) {
	references := newProfileReferences(&profileReferenceFixture{})
	_, err := references.AssertRetirable(
		context.Background(), platformtx.TenantScope{},
		iammodule.ProfileRetirementTarget{
			Resource: "unknown",
			FieldID: [16]byte{
				1, 2, 3, 4, 5, 6, 0x40, 8, 0x80, 10, 11, 12, 13, 14, 15, 16,
			},
			ExpectedFieldRevision: 1,
		},
	)
	if err == nil {
		t.Fatal("unknown resource was accepted")
	}
}
