package serviceprincipal

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAuthorizeRequiresExactPinnedGrant(t *testing.T) {
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	principal := Principal{
		ID: uuid.New(), Kind: KindWorkflow, State: PrincipalActive,
		AuthorizationVersion: 3, Revision: 2,
	}
	capability, err := ParseCapabilityKey("forms.record.update")
	if err != nil {
		t.Fatal(err)
	}
	scope := ResourceScope{
		Kind: ScopeResource, Owner: "forms", ResourceKind: "form", ResourceID: uuid.New(),
	}
	grant := GrantRevision{
		GrantID: uuid.New(), Revision: 2, PreviousRevision: 1,
		PrincipalID: principal.ID, Capability: capability, Scope: scope,
		ValidFrom: now.Add(-time.Hour), State: GrantActive,
		ManifestSHA256: [32]byte{1},
	}
	request := AuthorizationRequest{
		Capability: capability, Scope: scope, GrantID: grant.GrantID,
		GrantRevision: 2, Manifest: [32]byte{1}, At: now,
	}
	if err := Evaluate(principal, grant, request); err != nil {
		t.Fatal(err)
	}
	request.GrantRevision = 1
	if err := Evaluate(principal, grant, request); !errors.Is(err, ErrGrantMismatch) {
		t.Fatalf("stale pinned revision error = %v", err)
	}
}

func TestCapabilityAndScopeRejectWildcards(t *testing.T) {
	if _, err := ParseCapabilityKey("forms.*"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("wildcard capability error = %v", err)
	}
	if err := (ResourceScope{Kind: ScopeTenant, Owner: "forms", ResourceKind: "*"}).Validate(); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("wildcard scope error = %v", err)
	}
}

func TestRetiredPrincipalCannotReactivate(t *testing.T) {
	if CanTransitionPrincipal(PrincipalRetired, PrincipalActive) {
		t.Fatal("retired principal became active")
	}
}

func TestGrantRequiresNonzeroManifest(t *testing.T) {
	capability, err := ParseCapabilityKey("workflow.instance.execute")
	if err != nil {
		t.Fatal(err)
	}
	grant := GrantRevision{
		GrantID: uuid.New(), Revision: 1, PrincipalID: uuid.New(),
		Capability: capability,
		Scope:      ResourceScope{Kind: ScopeTenant, Owner: "workflow", ResourceKind: "instance"},
		ValidFrom:  time.Now().UTC(), State: GrantActive,
	}
	if err := grant.Validate(); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("zero manifest error = %v", err)
	}
}
