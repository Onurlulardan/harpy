package transaction

import (
	"context"
	"errors"
	"testing"
)

func TestCredentialTenantScopeExposesOnlySealedEvidenceWhileActive(t *testing.T) {
	target, err := issueCredentialTarget(
		validCredentialEvidenceFor(CredentialRefresh, CredentialRefreshLocator),
	)
	if err != nil {
		t.Fatal(err)
	}
	state := &credentialScopeState{target: target}
	state.active.Store(true)
	scope := CredentialTenantScope{state: state}
	group, singleGroup := target.operation.singleGroup()
	if scope.TenantID() != target.tenantID ||
		scope.OperationID() != target.operationID ||
		scope.Purpose() != CredentialRefresh ||
		scope.LocatorKind() != CredentialRefreshLocator ||
		scope.LocatorBindingSHA256() != target.locatorBindingSHA256 ||
		scope.SourceRevision() != target.sourceRevision ||
		scope.SourceTokenID() != target.sourceTokenID ||
		scope.SourceFamilyID() != target.sourceFamilyID ||
		scope.SourceSelector() != target.sourceSelector ||
		!singleGroup || scope.AuthorityGroupGeneration() != group.generation ||
		scope.IAMGeneration() != target.operation.domains[0].generation ||
		scope.SystemResourceGeneration() != target.operation.domains[1].generation {
		t.Fatal("credential scope evidence changed")
	}
	if err := RequireCredentialPurpose(scope, CredentialRefresh); err != nil {
		t.Fatalf("exact purpose rejected: %v", err)
	}
	if err := RequireCredentialPurpose(scope, CredentialLogin); !errors.Is(err, ErrCredentialPurpose) {
		t.Fatalf("wrong purpose error = %v", err)
	}
	ctx := markCredentialScopeContext(context.Background(), scope)
	if !ContextOwnsCredentialScope(ctx, scope) || NetworkAllowed(ctx) {
		t.Fatal("credential scope context did not close the network boundary")
	}
	state.active.Store(false)
	assertClosedCredentialScope(t, scope)
}

func TestCredentialScopeContextRejectsForeignAndNilScopes(t *testing.T) {
	first := activeCredentialScopeForTest(t)
	second := activeCredentialScopeForTest(t)
	ctx := markCredentialScopeContext(context.Background(), first)
	if ContextOwnsCredentialScope(ctx, second) ||
		ContextOwnsCredentialScope(ctx, CredentialTenantScope{}) ||
		!nestedScope(ctx) {
		t.Fatal("credential scope identity was forgeable")
	}
}

func activeCredentialScopeForTest(t *testing.T) CredentialTenantScope {
	t.Helper()
	target, err := issueCredentialTarget(
		validCredentialEvidenceFor(CredentialLogin, CredentialLoginDirectory),
	)
	if err != nil {
		t.Fatal(err)
	}
	state := &credentialScopeState{target: target}
	state.active.Store(true)
	return CredentialTenantScope{state: state}
}

func assertClosedCredentialScope(t *testing.T, scope CredentialTenantScope) {
	t.Helper()
	if scope.TenantID() != [16]byte{} || scope.OperationID() != "" ||
		scope.Purpose() != "" || scope.LocatorKind() != "" ||
		scope.LocatorBindingSHA256() != [32]byte{} || scope.SourceRevision() != 0 ||
		scope.AuthorityGroupGeneration() != 0 || scope.IAMGeneration() != 0 ||
		scope.SystemResourceGeneration() != 0 {
		t.Fatal("closed credential scope leaked evidence")
	}
	if scope.SourceUserID() != [16]byte{} || scope.SourceIdentifierID() != [16]byte{} ||
		scope.SourceIdentifierKind() != "" || scope.SourceTokenID() != [16]byte{} ||
		scope.SourceFamilyID() != [16]byte{} || scope.SourceSelector() != [16]byte{} {
		t.Fatal("closed credential scope leaked exact source identity")
	}
	if err := RequireCredentialPurpose(scope, CredentialRefresh); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("closed purpose error = %v", err)
	}
}
