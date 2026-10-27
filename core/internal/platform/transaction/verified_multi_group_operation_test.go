package transaction

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestMultiGroupIssuerSealsCanonicalGroupsDomainsAndParticipants(t *testing.T) {
	groups, domains, participants := validMultiGroupAuthorityInput()
	operation, err := NewAuthorityIssuer().IssueMultiGroup(
		"workflow.forms.set-field", groups, domains, participants,
	)
	if err != nil || !operation.valid() {
		t.Fatalf("multi-group operation rejected: %+v %v", operation, err)
	}
	if len(operation.groups) != 2 || operation.groups[0].id != "billing_forms" ||
		operation.groups[1].id != "workflow" || len(operation.domains) != 2 ||
		operation.domains[0] != (domainExpectation{
			name: "billing", generation: 11, groupID: "billing_forms",
		}) || operation.domains[1] != (domainExpectation{
		name: "workflow", generation: 17, groupID: "workflow",
	}) {
		t.Fatalf("operation was not canonical: %+v", operation)
	}
	if !operation.allowsOwner("workflow") || operation.allowsOwner("billing") ||
		operation.allowsOwner("forms") {
		t.Fatalf("participant authority widened: %+v", operation.participants)
	}
	groups[0].ID, domains[0].Name, participants[0] = "work", "work", "billing"
	if operation.groups[1].id != "workflow" || operation.domains[1].name != "workflow" ||
		operation.participants[0] != "workflow" {
		t.Fatal("operation retained caller-owned slices")
	}
}

func TestMultiGroupIssuerRejectsInvalidAuthoritySets(t *testing.T) {
	validGroups, validDomains, validParticipants := validMultiGroupAuthorityInput()
	tests := []struct {
		name         string
		groups       []CommandAuthorityGroup
		domains      []CommandAuthorityDomain
		participants []string
	}{
		{"no groups", nil, validDomains, validParticipants},
		{"unknown group", []CommandAuthorityGroup{{ID: "unknown", Generation: 1}}, validDomains, validParticipants},
		{"duplicate group", append(validGroups, validGroups[0]), validDomains, validParticipants},
		{"mixed group generation", append(validGroups, CommandAuthorityGroup{ID: "workflow", Generation: 99}), validDomains, validParticipants},
		{"partial domain union", validGroups, validDomains[1:], validParticipants},
		{"extra domain", validGroups, append(validDomains, CommandAuthorityDomain{Name: "work", Generation: 1}), validParticipants},
		{"duplicate domain", validGroups, append(validDomains, validDomains[0]), validParticipants},
		{"mixed domain generation", validGroups, append(validDomains, CommandAuthorityDomain{Name: "workflow", Generation: 99}), validParticipants},
		{"no participants", validGroups, validDomains, nil},
		{"duplicate participant", validGroups, validDomains, []string{"forms", "forms"}},
		{"unknown participant", validGroups, validDomains, []string{"forms", "work"}},
		{"zero group generation", []CommandAuthorityGroup{{ID: "workflow"}}, []CommandAuthorityDomain{{Name: "workflow", Generation: 1}}, []string{"workflow"}},
		{"zero domain generation", []CommandAuthorityGroup{{ID: "workflow", Generation: 1}}, []CommandAuthorityDomain{{Name: "workflow"}}, []string{"workflow"}},
	}
	issuer := NewAuthorityIssuer()
	for _, test := range tests {
		if _, err := issuer.IssueMultiGroup(
			"workflow.forms.set-field", test.groups, test.domains, test.participants,
		); err == nil {
			t.Fatalf("%s authority was accepted", test.name)
		}
	}
}

func TestMultiGroupOperationTamperFailsClosed(t *testing.T) {
	groups, domains, participants := validMultiGroupAuthorityInput()
	operation, err := NewAuthorityIssuer().IssueMultiGroup(
		"workflow.forms.set-field", groups, domains, participants,
	)
	if err != nil {
		t.Fatal(err)
	}
	duplicateGroup := operation
	duplicateGroup.groups = append(append([]groupExpectation{}, operation.groups...), operation.groups[1])
	wrongOwningGroup := operation
	wrongOwningGroup.domains = append([]domainExpectation{}, operation.domains...)
	wrongOwningGroup.domains[0].groupID = "workflow"
	duplicateParticipant := operation
	duplicateParticipant.participants = []string{"forms", "forms"}
	if duplicateGroup.valid() || wrongOwningGroup.valid() || duplicateParticipant.valid() {
		t.Fatal("tampered multi-group operation remained valid")
	}
}

func TestMultiGroupOwnerAuthorityUsesContainingExactGroup(t *testing.T) {
	groups, domains, participants := validMultiGroupAuthorityInput()
	operation, err := NewAuthorityIssuer().IssueMultiGroup(
		"workflow.forms.set-field", groups, domains, participants,
	)
	if err != nil {
		t.Fatal(err)
	}
	state := &scopeState{tenantID: uuid.New(), operation: operation, mode: Command}
	state.active.Store(true)
	scope := TenantScope{state: state}
	workflow, err := MintOwnerAuthority(scope, "workflow")
	if err != nil || workflow.GroupID() != "workflow" || workflow.GroupGeneration() != 7 {
		t.Fatalf("Workflow authority mismatch: %+v %v", workflow, err)
	}
	if _, err := MintOwnerAuthority(scope, "billing"); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("non-participant Billing authority error = %v", err)
	}
}

func validMultiGroupAuthorityInput() (
	[]CommandAuthorityGroup, []CommandAuthorityDomain, []string,
) {
	return []CommandAuthorityGroup{
			{ID: "workflow", Generation: 7}, {ID: "billing_forms", Generation: 3},
		}, []CommandAuthorityDomain{
			{Name: "workflow", Generation: 17}, {Name: "billing", Generation: 11},
		}, []string{"workflow"}
}
