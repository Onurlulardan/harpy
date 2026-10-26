package platformiam

import (
	"testing"
	"time"
)

func TestOperatorRevokeRejectsStaleAndWrongPurposeStepUp(t *testing.T) {
	now := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	command := OperatorRevoke{
		TargetUserID: fixtureID(7), TargetExpectedAuthorizationVersion: 8,
		Capability: "platform.operations.health.read", Evidence: fixtureEvidence(),
	}
	command.StepUp = fixtureStepUp(now,
		OperatorPurpose("revoke", command.TargetUserID, command.Capability))
	if err := command.Validate(now); err != nil {
		t.Fatal(err)
	}
	command.StepUp.AuthTime = now.Add(-10*time.Minute - time.Nanosecond)
	if err := command.Validate(now); err == nil {
		t.Fatal("stale step-up accepted")
	}
	command.StepUp = fixtureStepUp(now,
		OperatorPurpose("grant", command.TargetUserID, command.Capability))
	if err := command.Validate(now); err == nil {
		t.Fatal("grant purpose accepted for revoke")
	}
}
