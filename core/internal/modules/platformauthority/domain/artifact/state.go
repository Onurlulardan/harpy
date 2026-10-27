package artifact

import "reflect"

type CanonicalState struct {
	Value  State
	Bytes  []byte
	SHA256 string
}

func CanonicalizeState(value State) (CanonicalState, error) {
	if !validState(value) {
		return CanonicalState{}, ErrInvalidDocument
	}
	bytes, err := canonicalValue(value)
	if err != nil {
		return CanonicalState{}, err
	}
	return CanonicalState{Value: value, Bytes: bytes, SHA256: digestHex(bytes)}, nil
}

func ParseState(document []byte) (CanonicalState, error) {
	var value State
	canonical, err := strictCanonicalDecode(document, &value)
	if err != nil || !validState(value) {
		return CanonicalState{}, ErrInvalidDocument
	}
	return CanonicalState{Value: value, Bytes: canonical, SHA256: digestHex(canonical)}, nil
}

func validState(value State) bool {
	if value.Schema != StateSchema || !validUUID(value.EnvironmentID) ||
		!validUUID(value.RegistryBindingID) || value.ProposalSignerGeneration < 1 ||
		!validKID(value.ProposalSignerKID) ||
		(value.DeploymentMode != ModeOnPremise && value.DeploymentMode != ModeSaaS) {
		return false
	}
	if value.ScopeKind == ScopeEnvironment {
		return validEnvironmentState(value)
	}
	if value.ScopeKind == ScopeTenant {
		return validTenantState(value)
	}
	return false
}

func validEnvironmentState(value State) bool {
	return value.ScopeID == value.EnvironmentID && value.TenantID == nil &&
		value.ProtectedAdminBindingSetSHA256 == nil &&
		value.OwnerBindingID != nil && validUUID(*value.OwnerBindingID) &&
		value.OwnerBindingGeneration != nil && *value.OwnerBindingGeneration > 0 &&
		value.PrimaryAdminBindingSetSHA256 != nil &&
		validDigest(*value.PrimaryAdminBindingSetSHA256) &&
		validOptionalDigest(value.RecoveryAnchorSetSHA256) &&
		validOptionalDigest(value.RecoveryTrustBundleSHA256)
}

func validTenantState(value State) bool {
	return value.DeploymentMode == ModeSaaS && validUUID(value.ScopeID) &&
		value.TenantID != nil && *value.TenantID == value.ScopeID &&
		value.ProtectedAdminBindingSetSHA256 != nil &&
		validDigest(*value.ProtectedAdminBindingSetSHA256) &&
		value.OwnerBindingID == nil && value.OwnerBindingGeneration == nil &&
		value.PrimaryAdminBindingSetSHA256 == nil &&
		value.RecoveryAnchorSetSHA256 == nil && value.RecoveryTrustBundleSHA256 == nil
}

func (state CanonicalState) Validate() error {
	parsed, err := ParseState(state.Bytes)
	if err != nil || parsed.SHA256 != state.SHA256 ||
		!reflect.DeepEqual(parsed.Value, state.Value) {
		return ErrInvalidDocument
	}
	return nil
}
