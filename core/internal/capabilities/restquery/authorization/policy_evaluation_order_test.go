package authorization

import (
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func TestOrderedPolicyRequestsUsesOwnerOrderAndOwnsInput(t *testing.T) {
	tenantID := uuid.MustParse("50000000-0000-4000-8000-000000000001")
	actorID := uuid.MustParse("60000000-0000-4000-8000-000000000001")
	queryIdentity := sha256.Sum256([]byte("same-query"))
	forms := policyRequestForOwner(t, "forms", tenantID, actorID, queryIdentity)
	workflow := policyRequestForOwner(t, "workflow", tenantID, actorID, queryIdentity)
	registry, err := NewPolicyRegistry([]PolicyRegistration{
		{Owner: "workflow", Port: &registryPolicyPort{}},
		{Owner: "forms", Port: &registryPolicyPort{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	input := []PolicyRequest{workflow, forms}
	ordered, err := orderedPolicyRequests(registry, tenantID, actorID, input)
	if err != nil || len(ordered) != 2 || ordered[0].owner != "forms" ||
		ordered[1].owner != "workflow" {
		t.Fatalf("ordered requests = %#v, error = %v", ordered, err)
	}
	if input[0].owner != "workflow" {
		t.Fatal("input order was mutated")
	}
}

func TestOrderedPolicyRequestsRejectsIdentityAndRegistryDrift(t *testing.T) {
	tenantID := uuid.MustParse("50000000-0000-4000-8000-000000000001")
	actorID := uuid.MustParse("60000000-0000-4000-8000-000000000001")
	queryIdentity := sha256.Sum256([]byte("same-query"))
	forms := policyRequestForOwner(t, "forms", tenantID, actorID, queryIdentity)
	registry, _ := NewPolicyRegistry([]PolicyRegistration{
		{Owner: "forms", Port: &registryPolicyPort{}},
	})
	cases := []struct {
		tenant   uuid.UUID
		actor    uuid.UUID
		requests []PolicyRequest
	}{
		{uuid.New(), actorID, []PolicyRequest{forms}},
		{tenantID, uuid.New(), []PolicyRequest{forms}},
		{tenantID, actorID, nil},
		{tenantID, actorID, []PolicyRequest{forms, forms}},
		{tenantID, actorID, []PolicyRequest{
			policyRequestForOwner(
				t, "workflow", tenantID, actorID, queryIdentity,
			),
		}},
	}
	for index, testCase := range cases {
		if _, err := orderedPolicyRequests(
			registry, testCase.tenant, testCase.actor, testCase.requests,
		); err == nil {
			t.Fatalf("invalid request batch %d was accepted", index)
		}
	}
}

func policyRequestForOwner(
	t *testing.T,
	owner string,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	queryIdentity [sha256.Size]byte,
) PolicyRequest {
	t.Helper()
	hash, _ := api.NewCatalogHash(strings.Repeat("b", 64))
	reference, err := api.NewCatalogReference(
		owner,
		"10000000-0000-4000-8000-000000000001",
		"20000000-0000-4000-8000-000000000001",
		hash,
	)
	if err != nil {
		t.Fatal(err)
	}
	path, _ := resolved.NewRequestPath("")
	requirement, err := newRequirement(
		RequirementResourceRead, reference, path, path, uuid.Nil, uuid.Nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := newPolicyRequest(
		tenantID, actorID, queryIdentity, owner, []Requirement{requirement},
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
