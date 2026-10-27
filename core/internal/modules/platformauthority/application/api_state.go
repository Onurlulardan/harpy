package application

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func artifactState(value api.State) artifact.State {
	return artifact.State{
		Schema: value.Schema, EnvironmentID: value.EnvironmentID,
		DeploymentMode:    artifact.DeploymentMode(value.DeploymentMode),
		RegistryBindingID: value.RegistryBindingID,
		ScopeKind:         artifact.ScopeKind(value.ScopeKind), ScopeID: value.ScopeID,
		ProposalSignerGeneration:       value.ProposalSignerGeneration,
		ProposalSignerKID:              value.ProposalSignerKID,
		OwnerBindingID:                 value.OwnerBindingID,
		OwnerBindingGeneration:         value.OwnerBindingGeneration,
		PrimaryAdminBindingSetSHA256:   value.PrimaryAdminBindingSetSHA256,
		RecoveryAnchorSetSHA256:        value.RecoveryAnchorSetSHA256,
		RecoveryTrustBundleSHA256:      value.RecoveryTrustBundleSHA256,
		TenantID:                       value.TenantID,
		ProtectedAdminBindingSetSHA256: value.ProtectedAdminBindingSetSHA256,
	}
}

func artifactExpiry(value api.ExpiryPolicy) artifact.ExpiryPolicy {
	return artifact.ExpiryPolicy(value)
}
