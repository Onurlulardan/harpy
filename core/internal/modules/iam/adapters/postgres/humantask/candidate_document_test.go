package humantask

import (
	"testing"
	"time"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	"github.com/google/uuid"
)

func TestDecodeCandidateRejectsDisplayMetadataAndAcceptsStableFacts(t *testing.T) {
	actor := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	role := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	assignment := uuid.MustParse("10000000-0000-4000-8000-000000000003")
	request := humantaskapi.VerificationRequest{
		TenantID: [16]byte(uuid.MustParse("10000000-0000-4000-8000-000000000004")),
		ActorID:  [16]byte(actor), Candidate: humantaskapi.Candidate{
			Kind: "role", ID: [16]byte(role),
		}, Action: "claim", VerifyAt: time.Date(2026, 7, 18, 15, 0, 0, 0, time.UTC),
	}
	document := []byte(`{"actor_id":"` + actor.String() +
		`","authorization_version":4,"candidate_kind":"role","candidate_id":"` +
		role.String() + `","facts":[{"kind":"role","id":"` + role.String() +
		`","revision":2},{"kind":"role_assignment","id":"` +
		assignment.String() + `","revision":3}]}`)
	evidence, err := decodeCandidate(document, request)
	if err != nil || evidence.ActorID != request.ActorID ||
		evidence.AuthorizationVersion != 4 || evidence.Candidate != request.Candidate ||
		evidence.EvidenceSHA256 == [32]byte{} {
		t.Fatalf("candidate evidence = %+v, %v", evidence, err)
	}
	display := append(document[:len(document)-1], []byte(`,"display_name":"Admin"}`)...)
	if _, err = decodeCandidate(display, request); err == nil {
		t.Fatal("display metadata was accepted")
	}
}
