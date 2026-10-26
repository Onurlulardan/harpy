package transaction

import "github.com/google/uuid"

const iamCredentialEventGroup = "iam_system_resources"

// IAMCredentialEventAuthority is a sealed projection for pre-authentication
// IAM events. It never grants actor authority or a general tenant scope.
type IAMCredentialEventAuthority struct {
	state     *credentialScopeState
	subjectID uuid.UUID
}

func MintIAMCredentialEventAuthority(
	scope CredentialTenantScope,
	subjectID uuid.UUID,
) (IAMCredentialEventAuthority, error) {
	if !scope.valid() {
		return IAMCredentialEventAuthority{}, ErrScopeClosed
	}
	switch scope.Purpose() {
	case CredentialPasswordReset, CredentialInitialAdminSetup,
		CredentialEmailVerification,
		CredentialPortalSetup, CredentialInvitation:
	default:
		return IAMCredentialEventAuthority{}, ErrCredentialPurpose
	}
	group, valid := scope.state.target.operation.singleGroup()
	if subjectID == uuid.Nil || !valid || group.id != iamCredentialEventGroup {
		return IAMCredentialEventAuthority{}, ErrScopeOwner
	}
	authority := IAMCredentialEventAuthority{state: scope.state, subjectID: subjectID}
	if !authority.valid() {
		return IAMCredentialEventAuthority{}, ErrScopeOwner
	}
	return authority, nil
}

func (authority IAMCredentialEventAuthority) ScopeID() uuid.UUID {
	if !authority.valid() {
		return uuid.Nil
	}
	return authority.state.target.tenantID
}

func (authority IAMCredentialEventAuthority) GroupID() string {
	if !authority.valid() {
		return ""
	}
	group, _ := authority.state.target.operation.singleGroup()
	return group.id
}

func (authority IAMCredentialEventAuthority) GroupGeneration() int64 {
	if !authority.valid() {
		return 0
	}
	group, _ := authority.state.target.operation.singleGroup()
	return group.generation
}

func (authority IAMCredentialEventAuthority) IAMGeneration() int64 {
	if !authority.valid() {
		return 0
	}
	return CredentialTenantScope{state: authority.state}.IAMGeneration()
}

func (authority IAMCredentialEventAuthority) SubjectID() uuid.UUID {
	if !authority.valid() {
		return uuid.Nil
	}
	return authority.subjectID
}

func (authority IAMCredentialEventAuthority) OperationID() string {
	if !authority.valid() {
		return ""
	}
	return authority.state.target.operationID
}

func (authority IAMCredentialEventAuthority) valid() bool {
	if authority.state == nil || !authority.state.active.Load() ||
		authority.subjectID == uuid.Nil || !authority.state.target.valid() {
		return false
	}
	group, singleGroup := authority.state.target.operation.singleGroup()
	if !singleGroup || group.id != iamCredentialEventGroup {
		return false
	}
	switch authority.state.target.purpose {
	case CredentialPasswordReset, CredentialInitialAdminSetup,
		CredentialEmailVerification,
		CredentialPortalSetup, CredentialInvitation:
		return true
	default:
		return false
	}
}
