package postgres

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

type reauthenticationCredentialRow struct {
	credential                                       authentication.ReauthenticationCredential
	challengeHMAC, contextSHA, purposeSHA, originSHA []byte
	accessMode                                       string
	portalID                                         *uuid.UUID
	portalState                                      *string
	portalRevision                                   *int64
	passwordVersion, passwordRevision                *int64
	passwordAlgorithm                                *string
	passwordSalt, passwordTag                        []byte
	passwordMemory                                   *int32
	passwordIterations, passwordParallelism          *int16
	passwordLegacy                                   *string
	passwordCreatedAt                                *time.Time
	authenticatorID                                  *uuid.UUID
	authenticatorState, authenticatorKind            *string
	authenticatorRevision                            *int64
	webCredentialID, webPublicKey, webRPIDSHA        []byte
	webRPID                                          *string
	webOriginPolicy, webSignCount, webRevision       *int64
	webBackupEligible, webBackupState, webUV         *bool
	totpProfile, totpCipherSuite, totpKMSKeyID       *string
	totpNonce, totpCiphertext, totpTag, totpAAD      []byte
	totpLastStep, totpRevision                       *int64
}

func (row *reauthenticationCredentialRow) targets() []any {
	credential := &row.credential
	return []any{
		&credential.ChallengeRevision, &credential.AttemptCount,
		&credential.ChallengeState, &credential.HMACKeyID,
		&row.challengeHMAC, &row.contextSHA, &row.purposeSHA, &row.originSHA,
		&credential.ActionPolicyVersion, &credential.ExpiresAt,
		&credential.UserRevision, &credential.TokenVersion,
		&credential.AuthorizationVersion, &credential.UserStatus,
		&credential.IdentityKind, &credential.SessionRevision,
		&credential.SessionState, &credential.SessionAbsoluteExpiresAt,
		&credential.AccessModeID, &row.accessMode, &credential.AccessModeState,
		&credential.AccessModeRevision, &row.portalID, &row.portalState,
		&row.portalRevision, &credential.FamilyID, &credential.FamilyRevision,
		&credential.FamilyState, &credential.FamilyAbsoluteExpiresAt,
		&credential.TokenID, &credential.TokenRevision, &credential.TokenState,
		&row.passwordVersion, &row.passwordRevision, &row.passwordAlgorithm,
		&row.passwordSalt, &row.passwordTag, &row.passwordMemory,
		&row.passwordIterations, &row.passwordParallelism, &row.passwordLegacy,
		&row.passwordCreatedAt, &row.authenticatorID, &row.authenticatorState,
		&row.authenticatorKind, &row.authenticatorRevision,
		&row.webCredentialID, &row.webPublicKey, &row.webRPID, &row.webRPIDSHA,
		&row.webOriginPolicy, &row.webSignCount, &row.webBackupEligible,
		&row.webBackupState, &row.webUV, &row.webRevision, &row.totpProfile,
		&row.totpCipherSuite, &row.totpKMSKeyID, &row.totpNonce,
		&row.totpCiphertext, &row.totpTag, &row.totpAAD, &row.totpLastStep,
		&row.totpRevision,
	}
}

func (row *reauthenticationCredentialRow) clear() {
	if row == nil {
		return
	}
	for _, value := range [][]byte{
		row.challengeHMAC, row.contextSHA, row.purposeSHA, row.originSHA,
		row.passwordSalt, row.passwordTag, row.webCredentialID,
		row.webPublicKey, row.webRPIDSHA, row.totpNonce, row.totpCiphertext,
		row.totpTag, row.totpAAD,
	} {
		clear(value)
	}
}
