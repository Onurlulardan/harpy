package transaction

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestIAMCredentialEventAuthorityAcceptsOnlyCredentialEventPurposes(t *testing.T) {
	subjectID := uuid.MustParse("ab000000-0000-4000-8000-000000000001")
	for _, purpose := range []CredentialPurpose{
		CredentialPasswordReset, CredentialEmailVerification,
		CredentialPortalSetup, CredentialInvitation,
	} {
		scope := credentialScopeForPurpose(t, purpose)
		authority, err := MintIAMCredentialEventAuthority(scope, subjectID)
		if err != nil || authority.ScopeID() != scope.TenantID() ||
			authority.GroupID() != credentialAuthorityGroup ||
			authority.GroupGeneration() != scope.AuthorityGroupGeneration() ||
			authority.IAMGeneration() != scope.IAMGeneration() ||
			authority.SubjectID() != subjectID || authority.OperationID() != scope.OperationID() {
			t.Fatalf("purpose %q authority = %+v, %v", purpose, authority, err)
		}
	}
}

func TestIAMCredentialEventAuthorityRejectsClosedWrongPurposeAndNilSubject(t *testing.T) {
	subjectID := uuid.MustParse("ab000000-0000-4000-8000-000000000002")
	if _, err := MintIAMCredentialEventAuthority(
		CredentialTenantScope{}, subjectID,
	); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("closed scope error = %v", err)
	}
	if _, err := MintIAMCredentialEventAuthority(
		credentialScopeForPurpose(t, CredentialLogin), subjectID,
	); !errors.Is(err, ErrCredentialPurpose) {
		t.Fatalf("wrong purpose error = %v", err)
	}
	if _, err := MintIAMCredentialEventAuthority(
		credentialScopeForPurpose(t, CredentialPasswordReset), uuid.Nil,
	); !errors.Is(err, ErrScopeOwner) {
		t.Fatalf("nil subject error = %v", err)
	}
}

func TestIAMCredentialEventAuthorityClosesWithScope(t *testing.T) {
	scope := credentialScopeForPurpose(t, CredentialEmailVerification)
	authority, err := MintIAMCredentialEventAuthority(
		scope, uuid.MustParse("ab000000-0000-4000-8000-000000000003"),
	)
	if err != nil {
		t.Fatal(err)
	}
	scope.state.active.Store(false)
	if authority.ScopeID() != uuid.Nil || authority.GroupID() != "" ||
		authority.GroupGeneration() != 0 || authority.IAMGeneration() != 0 ||
		authority.SubjectID() != uuid.Nil || authority.OperationID() != "" {
		t.Fatal("closed credential authority leaked evidence")
	}
}

func credentialScopeForPurpose(t *testing.T, purpose CredentialPurpose) CredentialTenantScope {
	t.Helper()
	kind := CredentialTenantRoute
	if purpose == CredentialLogin {
		kind = CredentialLoginDirectory
	}
	target, err := issueCredentialTarget(validCredentialEvidenceFor(purpose, kind))
	if err != nil {
		t.Fatal(err)
	}
	state := &credentialScopeState{target: target}
	state.active.Store(true)
	return CredentialTenantScope{state: state}
}
