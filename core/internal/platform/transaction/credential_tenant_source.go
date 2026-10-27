package transaction

import "github.com/google/uuid"

func (scope CredentialTenantScope) SourceUserID() uuid.UUID {
	if !scope.valid() || scope.LocatorKind() != CredentialLoginDirectory {
		return uuid.Nil
	}
	return scope.state.target.sourceUserID
}

func (scope CredentialTenantScope) SourceIdentifierID() uuid.UUID {
	if !scope.valid() || scope.LocatorKind() != CredentialLoginDirectory {
		return uuid.Nil
	}
	return scope.state.target.sourceIdentifierID
}

func (scope CredentialTenantScope) SourceIdentifierKind() CredentialIdentifierKind {
	if !scope.valid() || scope.LocatorKind() != CredentialLoginDirectory {
		return ""
	}
	return scope.state.target.sourceIdentifierKind
}

func (scope CredentialTenantScope) SourceTokenID() uuid.UUID {
	if !scope.valid() || scope.LocatorKind() != CredentialRefreshLocator {
		return uuid.Nil
	}
	return scope.state.target.sourceTokenID
}

func (scope CredentialTenantScope) SourceFamilyID() uuid.UUID {
	if !scope.valid() || scope.LocatorKind() != CredentialRefreshLocator {
		return uuid.Nil
	}
	return scope.state.target.sourceFamilyID
}

func (scope CredentialTenantScope) SourceSelector() [16]byte {
	if !scope.valid() || scope.LocatorKind() != CredentialRefreshLocator {
		return [16]byte{}
	}
	return scope.state.target.sourceSelector
}
