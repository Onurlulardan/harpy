package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) TOTPEnrollmentLookup(ctx context.Context, scope tenantScope,
	query lifecycle.TOTPReadQuery) (lifecycle.TOTPEnrollment, error) {
	if !snapshotUsable(ctx, scope) || scope.ActorID() != query.ActorID {
		return lifecycle.TOTPEnrollment{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.TOTPEnrollment{}, err
	}
	var result lifecycle.TOTPEnrollment
	var aad, challengeHMAC, challengeContext []byte
	err = bound.database.QueryRow(ctx, totpReadSQL, query.AuthenticatorID,
		query.ChallengeID, query.ActorID, query.PermissionID,
		query.ExpectedTokenVersion, query.ExpectedAuthorizationVersion).Scan(
		&result.UserID, &result.KMSKeyID, &result.Nonce, &result.Ciphertext,
		&result.AuthenticationTag, &aad, &result.AuthenticatorRevision,
		&result.ChallengeHMACKeyID, &challengeHMAC, &challengeContext,
		&result.ChallengeRevision, &result.ExpiresAt)
	if err != nil {
		return lifecycle.TOTPEnrollment{}, mapError(err)
	}
	if len(aad) != 32 || len(challengeHMAC) != 32 || len(challengeContext) != 32 {
		clear(aad)
		clear(challengeHMAC)
		clear(challengeContext)
		return lifecycle.TOTPEnrollment{}, lifecycle.ErrUnavailable
	}
	copy(result.AADSHA256[:], aad)
	copy(result.ChallengeHMAC[:], challengeHMAC)
	copy(result.ContextSHA256[:], challengeContext)
	clear(aad)
	clear(challengeHMAC)
	clear(challengeContext)
	return result, nil
}
