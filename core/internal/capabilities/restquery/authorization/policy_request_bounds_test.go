package authorization

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestPolicyRequestRequiresExactOwnerAndResourceRead(t *testing.T) {
	tenantID := uuid.MustParse("50000000-0000-4000-8000-000000000001")
	actorID := uuid.MustParse("60000000-0000-4000-8000-000000000001")
	queryIdentity := sha256.Sum256([]byte("query"))
	if _, err := newPolicyRequest(
		tenantID, actorID, queryIdentity, "iam", policyRequirements(t),
	); err == nil {
		t.Fatal("catalog owner mismatch was accepted")
	}
	fieldOnly := policyRequirements(t)[:1]
	if _, err := newPolicyRequest(
		tenantID, actorID, queryIdentity, "forms", fieldOnly,
	); err == nil {
		t.Fatal("policy request without resource read was accepted")
	}
	if _, err := newPolicyRequest(
		uuid.Nil, actorID, queryIdentity, "forms", policyRequirements(t),
	); err == nil {
		t.Fatal("zero tenant identity was accepted")
	}
}

func TestPolicyRequestRequirementBound(t *testing.T) {
	requirements := make([]Requirement, 0, maximumPolicyRequirements+1)
	requirements = append(requirements, mustRequirement(
		t, RequirementResourceRead, "", "", uuid.Nil, uuid.Nil,
	))
	for index := 1; index < maximumPolicyRequirements; index++ {
		requirements = append(requirements, mustRequirement(
			t, RequirementFieldSelect, "", fmt.Sprintf("/select/%d", index),
			uuid.MustParse(testFieldID), uuid.Nil,
		))
	}
	tenantID := uuid.MustParse("50000000-0000-4000-8000-000000000001")
	actorID := uuid.MustParse("60000000-0000-4000-8000-000000000001")
	queryIdentity := sha256.Sum256([]byte("query"))
	if _, err := newPolicyRequest(
		tenantID, actorID, queryIdentity, "forms", requirements,
	); err != nil {
		t.Fatalf("exact policy requirement bound: %v", err)
	}
	requirements = append(requirements, mustRequirement(
		t, RequirementFieldSelect, "", "/select/overflow",
		uuid.MustParse(testFieldID), uuid.Nil,
	))
	_, err := newPolicyRequest(
		tenantID, actorID, queryIdentity, "forms", requirements,
	)
	if err == nil {
		t.Fatal("policy requirement bound was exceeded")
	}
}
