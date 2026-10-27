package projections

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

func TestProjectionFingerprintBindsVisiblePayload(t *testing.T) {
	base := projectionFixture()
	if !validProjectionInput(base, true) {
		t.Fatal("valid projection fixture was rejected")
	}
	first := projectionFingerprint(base, true)
	changed := base
	changed.Summary = "changed"
	if first == projectionFingerprint(changed, true) {
		t.Fatal("summary was absent from projection fingerprint")
	}
	changed = base
	changed.VisibleTo = []domain.Principal{{Kind: domain.PrincipalRole,
		ID: domain.ID(uuid.New())}}
	if first == projectionFingerprint(changed, true) {
		t.Fatal("principal visibility was absent from projection fingerprint")
	}
}

func TestProjectionFingerprintCanonicalizesPrincipalOrder(t *testing.T) {
	base := projectionFixture()
	base.VisibleTo = append(base.VisibleTo, domain.Principal{
		Kind: domain.PrincipalRole, ID: domain.ID(uuid.New())})
	reversed := base
	reversed.VisibleTo = []domain.Principal{base.VisibleTo[1], base.VisibleTo[0]}
	if projectionFingerprint(base, true) != projectionFingerprint(reversed, true) {
		t.Fatal("equivalent principal sets produced different fingerprints")
	}
}

func projectionFixture() ports.ProjectionItem {
	due := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	return ports.ProjectionItem{
		ID: domain.ID(uuid.New()), SourceID: domain.ID(uuid.New()),
		SourceVersionID: domain.ID(uuid.New()), SourceOwner: domain.SourceWorkflow,
		SourceKind: "HUMAN_TASK", Surface: "INBOX", Summary: "Review",
		VisibleTo: []domain.Principal{{Kind: domain.PrincipalUser,
			ID: domain.ID(uuid.New())}}, SourceRevision: 2, ReferenceEpoch: 3,
		PermissionRevision: 4, EventSequence: 5, DueAt: &due,
	}
}
