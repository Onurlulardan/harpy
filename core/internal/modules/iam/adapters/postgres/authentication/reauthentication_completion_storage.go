package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

func (bound *boundStore) performReauthenticationCompletion(
	ctx context.Context,
	scope authentication.CredentialScope,
	completion authentication.ReauthenticationCompletion,
) (reauthenticationCompletionRow, error) {
	portalID, portalRevision := portalArguments(
		completion.PortalBindingID, completion.ExpectedPortalBindingRevision,
	)
	var authenticatorID any
	if completion.AuthenticatorID != uuid.Nil {
		authenticatorID = completion.AuthenticatorID
	}
	binding := scope.LocatorBindingSHA256()
	purpose := completion.PurposeSHA256
	contextSHA := completion.ContextSHA256
	selector := completion.SuccessorSelector
	secretHMAC := completion.SuccessorSecretHMAC
	var row reauthenticationCompletionRow
	err := bound.database.QueryRow(ctx, completeReauthenticationSQL,
		scope.TenantID(), binding[:], completion.ChallengeID,
		completion.ExpectedChallengeRevision, completion.ExpectedAttemptCount,
		completion.UserID, completion.SessionID, completion.ExpectedUserRevision,
		completion.ExpectedSessionRevision, completion.ExpectedTokenVersion,
		completion.ExpectedAuthorizationVersion, completion.AccessModeID,
		string(completion.AccessMode), completion.ExpectedAccessModeRevision,
		portalID, portalRevision, completion.FamilyID, completion.TokenID,
		completion.ExpectedFamilyRevision, completion.ExpectedTokenRevision,
		authenticatorID, completion.ExpectedAuthenticatorRevision,
		completion.ExpectedSubtypeRevision,
		completion.ExpectedPasswordCredentialVersion,
		completion.ExpectedPasswordCredentialRevision, purpose[:], contextSHA[:],
		string(completion.Proof.Method), completion.Proof.WebAuthnSignCount,
		completion.Proof.WebAuthnBackupEligible,
		completion.Proof.WebAuthnBackupState, completion.Proof.TOTPTimeStep,
		completion.SuccessorID, selector[:], completion.SuccessorHashKeyID,
		secretHMAC[:],
	).Scan(
		&row.completed.IssuedAt, &row.completed.ChallengeRevision,
		&row.completed.FamilyRevision, &row.completed.SessionRevision,
		&row.completed.ConsumedTokenRevision,
		&row.completed.SuccessorTokenRevision, &row.advance.selector,
		&row.advance.expected, &row.advance.revision, &row.advance.sourceSHA256,
		&row.advance.state, &row.advance.projectedAt, &row.advance.retainUntil,
		&row.successor.selector, &row.successor.revision,
		&row.successor.sourceSHA256, &row.successor.state,
		&row.successor.issuedAt, &row.successor.projectedAt,
		&row.successor.familyExpires, &row.successor.retainUntil,
	)
	if err != nil || !row.valid() {
		row.clear()
		return reauthenticationCompletionRow{}, errAuthenticationStorage
	}
	return row, nil
}
