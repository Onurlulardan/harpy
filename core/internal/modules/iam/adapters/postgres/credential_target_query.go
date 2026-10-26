package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (resolver *credentialTargetResolver) queryLoginTarget(
	ctx context.Context,
	tenantID uuid.UUID,
	kind string,
	keyID string,
	key []byte,
) (credentialTargetRow, error) {
	var row credentialTargetRow
	err := resolver.database.QueryRow(
		ctx, resolveLoginCredentialTargetSQL, tenantID, kind, keyID, key,
	).Scan(
		&row.tenantID, &row.userID, &row.identifierID, &row.identifierKind,
		&row.sourceRevision, &row.sourceSHA256,
		&row.groupGen, &row.iamGen, &row.layoutGen,
	)
	if err != nil || !row.validLogin() {
		return credentialTargetRow{}, errCredentialTargetUnavailable
	}
	return row, nil
}

func (resolver *credentialTargetResolver) queryRefreshTarget(
	ctx context.Context,
	selector []byte,
) (credentialTargetRow, error) {
	var row credentialTargetRow
	var storedSelector []byte
	err := resolver.database.QueryRow(
		ctx, resolveRefreshCredentialTargetSQL, selector,
	).Scan(
		&row.tenantID, &row.tokenID, &row.familyID, &storedSelector,
		&row.sourceRevision, &row.sourceSHA256,
		&row.groupGen, &row.iamGen, &row.layoutGen,
	)
	if len(storedSelector) == len(row.selector) {
		copy(row.selector[:], storedSelector)
	}
	clear(storedSelector)
	if err != nil || !row.validRefresh() {
		return credentialTargetRow{}, errCredentialTargetUnavailable
	}
	return row, nil
}

func (resolver *credentialTargetResolver) queryTenantTarget(
	ctx context.Context,
	tenantID uuid.UUID,
	binding []byte,
) (credentialTargetRow, error) {
	var row credentialTargetRow
	err := resolver.database.QueryRow(
		ctx, resolveTenantCredentialTargetSQL, tenantID, binding,
	).Scan(
		&row.tenantID, &row.sourceRevision, &row.sourceSHA256,
		&row.groupGen, &row.iamGen, &row.layoutGen,
	)
	if err != nil || !row.validBase() {
		return credentialTargetRow{}, errCredentialTargetUnavailable
	}
	return row, nil
}

func (row credentialTargetRow) validBase() bool {
	return row.tenantID != uuid.Nil && row.sourceRevision > 0 &&
		len(row.sourceSHA256) == 32 && row.groupGen > 0 &&
		row.iamGen > 0 && row.layoutGen > 0
}

func (row credentialTargetRow) validLogin() bool {
	return row.validBase() && row.userID != uuid.Nil && row.identifierID != uuid.Nil &&
		(row.identifierKind == "email_v1" || row.identifierKind == "portal_v1")
}

func (row credentialTargetRow) validRefresh() bool {
	return row.validBase() && row.tokenID != uuid.Nil && row.familyID != uuid.Nil &&
		row.selector != [16]byte{}
}
