package authorization

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

func TestGroupPolicyRequestsBuildsOneCanonicalRequestPerOwner(t *testing.T) {
	tenantID := uuid.MustParse("50000000-0000-4000-8000-000000000001")
	actorID := uuid.MustParse("60000000-0000-4000-8000-000000000001")
	queryIdentity := sha256.Sum256([]byte("grouped-query"))
	forms := policyRequestForOwner(
		t, "forms", tenantID, actorID, queryIdentity,
	).requirements[0]
	workflow := policyRequestForOwner(
		t, "workflow", tenantID, actorID, queryIdentity,
	).requirements[0]
	input := []Requirement{workflow, forms}
	requests, err := groupPolicyRequests(
		tenantID, actorID, queryIdentity, input,
	)
	if err != nil || len(requests) != 2 || requests[0].owner != "forms" ||
		requests[1].owner != "workflow" {
		t.Fatalf("grouped requests = %#v, error = %v", requests, err)
	}
	if input[0].catalog != workflow.catalog {
		t.Fatal("input requirements were mutated")
	}
}

func TestGroupPolicyRequestsRejectsMissingResourceAndDuplicateRequirement(t *testing.T) {
	request := mustPolicyRequest(t)
	resource := request.requirements[0]
	for _, candidate := range request.requirements {
		if candidate.kind == RequirementResourceRead {
			resource = candidate
			break
		}
	}
	fieldOnly, err := newRequirement(
		RequirementFieldSelect, resource.catalog, resource.resourcePath,
		requirementPath(t, "/select/0"), uuid.MustParse(testFieldID), uuid.Nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	cases := [][]Requirement{
		nil,
		{fieldOnly},
		{resource, resource},
		{{}},
	}
	for index, requirements := range cases {
		if _, err := groupPolicyRequests(
			request.tenantID, request.actorID, request.queryIdentity, requirements,
		); err == nil {
			t.Fatalf("invalid requirement set %d was accepted", index)
		}
	}
}
