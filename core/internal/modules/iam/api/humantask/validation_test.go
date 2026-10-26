package humantask

import (
	"testing"
	"time"
)

func TestHumanTaskEvidenceAcceptsAdoptedRFCIdentities(t *testing.T) {
	now := time.Date(2026, 7, 18, 15, 0, 0, 123000, time.UTC)
	request := VerificationRequest{
		TenantID: stableID(1), ActorID: stableID(2),
		Candidate: Candidate{Kind: "role", ID: stableID(3)},
		Action:    "decide", VerifyAt: now,
	}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
	evidence := VerificationEvidence{
		TenantID: request.TenantID, ActorID: request.ActorID,
		AuthorizationVersion: 4, Candidate: request.Candidate, Action: request.Action,
		PermissionID:   stableID(4),
		EvidenceSHA256: [32]byte{1}, VerifiedAt: now,
	}
	if !evidence.Matches(request) {
		t.Fatal("exact adopted evidence did not match")
	}
}

func stableID(last byte) [16]byte {
	value := [16]byte{0x10, 0, 0, 0, 0, 0, 0x40, 0, 0x80}
	value[15] = last
	return value
}

func TestHumanTaskEvidenceRejectsDisplayOrUnknownSemantics(t *testing.T) {
	request := VerificationRequest{Action: "display_name", VerifyAt: time.Now()}
	if request.Validate() == nil {
		t.Fatal("unknown action was accepted")
	}
}
