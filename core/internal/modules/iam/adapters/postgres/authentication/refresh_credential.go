package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

func (store *store) RefreshCredential(
	ctx context.Context,
	scope authentication.CredentialScope,
) (authentication.RefreshCredential, error) {
	bound, err := store.bindRefresh(ctx, scope)
	if err != nil {
		return authentication.RefreshCredential{}, err
	}
	return bound.refreshCredential(ctx, scope)
}

func (bound *boundStore) refreshCredential(
	ctx context.Context,
	scope authentication.CredentialScope,
) (authentication.RefreshCredential, error) {
	var result authentication.RefreshCredential
	var accessMode string
	var portalID *uuid.UUID
	var portalState *string
	var portalRevision *int64
	var secretHMAC, stepUp, securityEpoch []byte
	selector := scope.SourceSelector()
	binding := scope.LocatorBindingSHA256()
	err := bound.database.QueryRow(ctx, refreshCredentialSQL,
		scope.SourceTokenID(), scope.SourceFamilyID(), selector[:],
		scope.SourceRevision(), binding[:],
	).Scan(
		&result.TokenID, &result.TokenRevision, &result.TokenState,
		&result.FamilyID, &result.FamilyRevision, &result.FamilyState,
		&result.SessionID, &result.SessionRevision, &result.SessionState,
		&result.UserID, &result.UserRevision, &result.UserStatus,
		&result.IdentityKind, &result.TokenVersion, &result.AuthorizationVersion,
		&result.AccessModeID, &accessMode, &result.AccessModeState,
		&result.AccessModeRevision, &portalID, &portalState, &portalRevision,
		&result.HashKeyID, &secretHMAC, &result.Sequence, &result.AuthTime,
		&result.ACR, &result.AMRPassword, &result.AMRWebAuthn, &result.AMRTOTP,
		&result.AMRRecovery, &stepUp, &result.IdleExpiresAt,
		&result.AbsoluteExpiresAt, &result.RetainUntil,
		&result.PlatformAuthorizationClock, &securityEpoch,
	)
	if err != nil || len(secretHMAC) != len(result.SecretHMAC) ||
		(len(stepUp) != 0 && len(stepUp) != len(result.StepUpPurposeSHA256)) ||
		len(securityEpoch) != len(result.SecurityEpoch) {
		clear(secretHMAC)
		clear(stepUp)
		clear(securityEpoch)
		return authentication.RefreshCredential{}, errAuthenticationStorage
	}
	result.AccessMode = authentication.AccessMode(accessMode)
	result.Selector = selector
	copy(result.SecretHMAC[:], secretHMAC)
	clear(secretHMAC)
	if portalID != nil {
		result.PortalBindingID = *portalID
	}
	if portalState != nil {
		result.PortalBindingState = *portalState
	}
	if portalRevision != nil {
		result.PortalBindingRevision = *portalRevision
	}
	if len(stepUp) != 0 {
		copy(result.StepUpPurposeSHA256[:], stepUp)
		result.HasStepUpPurpose = true
	}
	clear(stepUp)
	copy(result.SecurityEpoch[:], securityEpoch)
	clear(securityEpoch)
	locatorSHA256, err := authentication.RefreshLocatorSourceSHA256(
		result.TokenID, result.FamilyID, result.Selector,
		result.TokenRevision, result.TokenState, result.SecretHMAC,
	)
	if err != nil {
		return authentication.RefreshCredential{}, errAuthenticationStorage
	}
	result.LocatorSourceSHA256 = locatorSHA256
	return result, nil
}
