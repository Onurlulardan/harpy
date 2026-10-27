package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

func (store *store) LookupCredential(ctx context.Context,
	scope credentialScope, lookup lifecycle.CredentialLookup,
) (lifecycle.CredentialRecord, error) {
	purpose, valid := lifecycleCredentialPurpose(lookup.Purpose)
	if lookup.Validate() != nil || !valid || !credentialUsable(ctx, scope, purpose) ||
		scope.LocatorBindingSHA256() != lookup.BindingSHA256 {
		return lifecycle.CredentialRecord{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindCredential(scope)
	if err != nil {
		return lifecycle.CredentialRecord{}, err
	}
	var result lifecycle.CredentialRecord
	var portalBindingID *uuid.UUID
	var portalRevision *int64
	var selector, secretHMAC, binding []byte
	err = bound.database.QueryRow(ctx, credentialLookupSQL, string(lookup.Purpose),
		lookup.Selector[:], lookup.BindingSHA256[:]).Scan(
		&result.TokenID, &result.Purpose, &result.UserID, &result.IdentifierID,
		&result.IdentifierKind, &result.IdentifierRevision, &portalBindingID,
		&portalRevision, &selector, &result.HashKeyID, &secretHMAC, &binding,
		&result.Revision, &result.State, &result.ExpiresAt,
	)
	defer clear(selector)
	defer clear(secretHMAC)
	defer clear(binding)
	if err != nil {
		return lifecycle.CredentialRecord{}, mapError(err)
	}
	if len(selector) != 16 || len(secretHMAC) != 32 || len(binding) != 32 {
		return lifecycle.CredentialRecord{}, lifecycle.ErrUnavailable
	}
	copy(result.Selector[:], selector)
	copy(result.SecretHMAC[:], secretHMAC)
	copy(result.BindingSHA256[:], binding)
	if portalBindingID != nil {
		result.PortalBindingID = *portalBindingID
	}
	if portalRevision != nil {
		result.PortalRevision = *portalRevision
	}
	if result.Validate() != nil || result.Selector != lookup.Selector ||
		result.BindingSHA256 != lookup.BindingSHA256 {
		return lifecycle.CredentialRecord{}, lifecycle.ErrUnavailable
	}
	return result, nil
}

func lifecycleCredentialPurpose(
	purpose lifecycle.CredentialPurpose,
) (credentialPurpose, bool) {
	switch purpose {
	case lifecycle.PurposeEmailVerification:
		return credentialEmailVerification, true
	case lifecycle.PurposePasswordReset:
		return credentialPasswordReset, true
	case lifecycle.PurposeInitialAdminSetup:
		return credentialInitialAdminSetup, true
	case lifecycle.PurposePortalSetup:
		return credentialPortalSetup, true
	default:
		return "", false
	}
}
