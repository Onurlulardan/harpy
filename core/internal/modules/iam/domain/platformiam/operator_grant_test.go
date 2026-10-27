package platformiam

import (
	"testing"
	"time"
)

func TestOperatorGrantRequiresExactCapabilityAndPurposeBoundStepUp(t *testing.T) {
	now := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	command := OperatorGrant{
		SubjectID: fixtureID(3), RoleID: fixtureID(4),
		SubjectRoleBindingID: fixtureID(5), RolePermissionBindingID: fixtureID(6),
		TargetUserID: fixtureID(7), TargetExpectedAuthorizationVersion: 8,
		Capability: "platform.operations.health.read", Evidence: fixtureEvidence(),
	}
	command.StepUp = fixtureStepUp(now,
		OperatorPurpose("grant", command.TargetUserID, command.Capability))
	if err := command.Validate(now); err != nil {
		t.Fatal(err)
	}
	command.Capability = "platform.*"
	if err := command.Validate(now); err == nil {
		t.Fatal("wildcard capability accepted")
	}
	command.Capability = "platform.operations.health.read"
	command.TargetUserID = command.Evidence.ActorID
	if err := command.Validate(now); err == nil {
		t.Fatal("owner self-grant accepted")
	}
}
