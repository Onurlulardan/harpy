package humantask

import (
	"testing"
	"time"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	"github.com/google/uuid"
)

func TestEvidenceDigestBindsPermissionCandidateVersionAndTime(t *testing.T) {
	evidence := humantaskapi.VerificationEvidence{
		TenantID:             [16]byte(uuid.MustParse("10000000-0000-4000-8000-000000000001")),
		ActorID:              [16]byte(uuid.MustParse("10000000-0000-4000-8000-000000000002")),
		AuthorizationVersion: 3,
		Candidate: humantaskapi.Candidate{Kind: "role", ID: [16]byte(
			uuid.MustParse("10000000-0000-4000-8000-000000000003"))},
		Action:       "claim",
		PermissionID: [16]byte(uuid.MustParse("10000000-0000-4000-8000-000000000004")),
		VerifiedAt:   time.Date(2026, 7, 18, 15, 0, 0, 0, time.UTC),
	}
	candidate := [32]byte{1}
	first := evidenceDigest(evidence, candidate)
	evidence.AuthorizationVersion++
	if first == evidenceDigest(evidence, candidate) || first == [32]byte{} {
		t.Fatal("human task evidence digest did not bind current authorization")
	}
}
