package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) ReauthenticationCredential(
	ctx context.Context,
	scope authentication.CredentialScope,
	lookup authentication.ReauthenticationLookup,
) (authentication.ReauthenticationCredential, error) {
	bound, err := store.bindReauthentication(ctx, scope)
	if err != nil {
		return authentication.ReauthenticationCredential{}, err
	}
	portalID, _ := portalArguments(lookup.PortalBindingID, 1)
	binding := scope.LocatorBindingSHA256()
	purpose := lookup.PurposeSHA256
	origin := lookup.OriginSHA256
	contextSHA := lookup.ContextSHA256
	var row reauthenticationCredentialRow
	err = bound.database.QueryRow(ctx, getReauthenticationCredentialSQL,
		scope.TenantID(), binding[:], lookup.ChallengeID, lookup.UserID,
		lookup.SessionID, lookup.ExpectedTokenVersion,
		lookup.ExpectedAuthorizationVersion, string(lookup.AccessMode), portalID,
		string(lookup.Method), purpose[:], lookup.ActionPolicyVersion,
		origin[:], contextSHA[:],
	).Scan(row.targets()...)
	if err != nil {
		row.clear()
		return authentication.ReauthenticationCredential{},
			errAuthenticationStorage
	}
	return row.decode(lookup)
}
