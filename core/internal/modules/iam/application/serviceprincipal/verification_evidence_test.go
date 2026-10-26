package serviceprincipal

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestVerificationEvidencePinsPrincipalAndGrantRevisions(t *testing.T) {
	capability, err := ParseCapabilityKey("workflow.instance.execute")
	if err != nil {
		t.Fatal(err)
	}
	principal := Principal{
		ID: uuid.New(), Kind: KindWorkflow, State: PrincipalActive,
		AuthorizationVersion: 9, Revision: 4,
	}
	grant := GrantRevision{
		GrantID: uuid.New(), Revision: 7, PrincipalID: principal.ID,
		Capability: capability,
		Scope: ResourceScope{
			Kind: ScopeResource, Owner: "workflow", ResourceKind: "definition",
			ResourceID: uuid.New(),
		},
		ManifestSHA256: [32]byte{1},
	}
	evidence := verifiedEvidence(
		principal, grant, time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC),
	)
	if err := evidence.Validate(); err != nil {
		t.Fatal(err)
	}
	if evidence.PrincipalID != principal.ID || evidence.PrincipalRevision != 4 ||
		evidence.AuthorizationVersion != 9 || evidence.GrantRevision != 7 {
		t.Fatalf("unexpected evidence: %#v", evidence)
	}
	evidence.PrincipalRevision = 0
	if !errors.Is(evidence.Validate(), ErrEvidenceMissing) {
		t.Fatal("revision-less evidence was accepted")
	}
}
