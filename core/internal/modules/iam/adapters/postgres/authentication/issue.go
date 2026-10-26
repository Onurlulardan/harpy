package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

func (store *store) Issue(
	ctx context.Context,
	scope authentication.CredentialScope,
	issue authentication.SessionIssue,
) (authentication.SessionIssued, error) {
	bound, err := store.bindLogin(ctx, scope)
	if err != nil {
		return authentication.SessionIssued{}, err
	}
	return bound.issue(ctx, scope, issue)
}

func (bound *boundStore) issue(
	ctx context.Context,
	scope authentication.CredentialScope,
	issue authentication.SessionIssue,
) (authentication.SessionIssued, error) {
	portalID, portalRevision := portalArguments(
		issue.PortalBindingID, issue.ExpectedPortalBindingRevision,
	)
	selector := issue.Selector
	secretHMAC := issue.SecretHMAC
	ipHMAC := issue.Device.IPHMAC
	var result authentication.SessionIssued
	var securityEpoch []byte
	err := bound.database.QueryRow(ctx, issueSessionSQL,
		issue.UserID, issue.ExpectedUserRevision, issue.ExpectedTokenVersion,
		issue.ExpectedAuthorizationVersion, issue.AccessModeID,
		string(issue.AccessMode), issue.ExpectedAccessModeRevision,
		portalID, portalRevision, issue.SessionID, issue.Device.Kind,
		issue.Device.BrowserFamily, issue.Device.OSFamily, issue.Device.IPKeyID,
		ipHMAC[:], issue.FamilyID, issue.TokenID, selector[:],
		issue.HashKeyID, secretHMAC[:],
	).Scan(
		&result.SessionRevision, &result.FamilyRevision, &result.TokenRevision,
		&result.IssuedAt, &result.IdleExpiresAt, &result.AbsoluteExpiresAt,
		&result.RetainUntil, &result.PlatformAuthorizationClock, &securityEpoch,
	)
	if err != nil || len(securityEpoch) != len(result.SecurityEpoch) {
		clear(securityEpoch)
		return authentication.SessionIssued{}, errAuthenticationStorage
	}
	copy(result.SecurityEpoch[:], securityEpoch)
	clear(securityEpoch)
	locatorSHA256, err := authentication.RefreshLocatorSourceSHA256(
		issue.TokenID, issue.FamilyID, selector, result.TokenRevision,
		"active", secretHMAC,
	)
	if err != nil {
		return authentication.SessionIssued{}, errAuthenticationStorage
	}
	_, err = bound.database.Execute(ctx, createRefreshLocatorSQL,
		selector[:], scope.TenantID(), issue.TokenID, issue.FamilyID,
		scope.AuthorityGroupGeneration(), scope.IAMGeneration(),
		result.TokenRevision, locatorSHA256[:], "active", result.IssuedAt,
		result.IssuedAt, result.AbsoluteExpiresAt, result.RetainUntil,
	)
	if err != nil {
		return authentication.SessionIssued{}, errAuthenticationStorage
	}
	return result, nil
}

func portalArguments(id uuid.UUID, revision int64) (any, any) {
	if id == uuid.Nil {
		return nil, nil
	}
	return id, revision
}
