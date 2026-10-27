package transaction

import (
	"testing"

	"github.com/google/uuid"
)

func TestCredentialTargetIssuerSealsExactPurposeAndAuthority(t *testing.T) {
	tests := []struct {
		purpose CredentialPurpose
		locator CredentialLocatorKind
	}{
		{CredentialLogin, CredentialLoginDirectory},
		{CredentialRefresh, CredentialRefreshLocator},
		{CredentialPasswordReset, CredentialTenantRoute},
		{CredentialInitialPassword, CredentialTenantRoute},
		{CredentialEmailVerification, CredentialTenantRoute},
		{CredentialPortalSetup, CredentialTenantRoute},
		{CredentialInvitation, CredentialTenantRoute},
		{CredentialReauthentication, CredentialSessionLocator},
	}
	for _, test := range tests {
		target, err := issueCredentialTarget(
			validCredentialEvidenceFor(test.purpose, test.locator),
		)
		if err != nil || !target.valid() {
			t.Fatalf("%s target rejected: %v", test.purpose, err)
		}
		group, singleGroup := target.operation.singleGroup()
		if !singleGroup || group.id != credentialAuthorityGroup ||
			group.generation != 7 ||
			len(target.operation.domains) != 2 ||
			target.operation.domains[0] != (domainExpectation{
				name: "iam", generation: 11, groupID: credentialAuthorityGroup,
			}) ||
			target.operation.domains[1] != (domainExpectation{
				name: "system_resource_layout", generation: 13,
				groupID: credentialAuthorityGroup,
			}) {
			t.Fatalf("%s authority was not sealed exactly: %+v", test.purpose, target.operation)
		}
	}
}

func TestCredentialTargetIssuerRejectsForgeableOrMismatchedEvidence(t *testing.T) {
	valid := validCredentialEvidenceFor(CredentialLogin, CredentialLoginDirectory)
	tests := []CredentialTargetEvidence{
		{},
		func() CredentialTargetEvidence { value := valid; value.OperationID = "bad operation"; return value }(),
		func() CredentialTargetEvidence { value := valid; value.TenantID = uuid.Nil; return value }(),
		func() CredentialTargetEvidence { value := valid; value.SourceUserID = uuid.Nil; return value }(),
		func() CredentialTargetEvidence { value := valid; value.SourceIdentifierID = uuid.Nil; return value }(),
		func() CredentialTargetEvidence { value := valid; value.SourceIdentifierKind = ""; return value }(),
		func() CredentialTargetEvidence {
			value := valid
			value.LocatorKind = CredentialRefreshLocator
			return value
		}(),
		func() CredentialTargetEvidence { value := valid; value.LocatorBindingSHA256 = [32]byte{}; return value }(),
		func() CredentialTargetEvidence { value := valid; value.SourceRevision = 0; return value }(),
		func() CredentialTargetEvidence { value := valid; value.GroupGeneration = 0; return value }(),
		func() CredentialTargetEvidence { value := valid; value.IAMGeneration = 0; return value }(),
		func() CredentialTargetEvidence {
			value := valid
			value.SystemResourceGeneration = 0
			return value
		}(),
	}
	for index, evidence := range tests {
		if _, err := issueCredentialTarget(evidence); err == nil {
			t.Fatalf("invalid evidence %d was accepted", index)
		}
	}
}

func TestCredentialTargetExactSourceCannotCrossLocatorKinds(t *testing.T) {
	login := validCredentialEvidenceFor(CredentialLogin, CredentialLoginDirectory)
	login.SourceTokenID, login.SourceFamilyID = uuid.New(), uuid.New()
	login.SourceSelector = [16]byte{9}
	if _, err := issueCredentialTarget(login); err == nil {
		t.Fatal("login evidence carrying refresh identity was accepted")
	}
	refresh := validCredentialEvidenceFor(CredentialRefresh, CredentialRefreshLocator)
	refresh.SourceUserID, refresh.SourceIdentifierID = uuid.New(), uuid.New()
	refresh.SourceIdentifierKind = CredentialPortalIdentifier
	if _, err := issueCredentialTarget(refresh); err == nil {
		t.Fatal("refresh evidence carrying login identity was accepted")
	}
}

func validCredentialEvidenceFor(
	purpose CredentialPurpose,
	locator CredentialLocatorKind,
) CredentialTargetEvidence {
	evidence := CredentialTargetEvidence{
		OperationID: "iam.credential.test", TenantID: uuid.New(),
		Purpose: purpose, LocatorKind: locator,
		LocatorBindingSHA256: [32]byte{1}, SourceRevision: 3,
		GroupGeneration: 7, IAMGeneration: 11, SystemResourceGeneration: 13,
	}
	switch locator {
	case CredentialLoginDirectory:
		evidence.SourceUserID = uuid.New()
		evidence.SourceIdentifierID = uuid.New()
		evidence.SourceIdentifierKind = CredentialEmailIdentifier
	case CredentialRefreshLocator:
		evidence.SourceTokenID = uuid.New()
		evidence.SourceFamilyID = uuid.New()
		evidence.SourceSelector = [16]byte{1}
	}
	return evidence
}
