package postgres

import (
	"context"
	"errors"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var errCredentialTargetUnavailable = errors.New("credential target is unavailable")

type credentialTargetResolver struct {
	database platformpostgres.BoundExecutor
}

type credentialTargetRow struct {
	tenantID       uuid.UUID
	userID         uuid.UUID
	identifierID   uuid.UUID
	identifierKind platformtx.CredentialIdentifierKind
	tokenID        uuid.UUID
	familyID       uuid.UUID
	requestID      uuid.UUID
	grantID        uuid.UUID
	selector       [16]byte
	sourceRevision int64
	sourceSHA256   []byte
	groupGen       int64
	iamGen         int64
	layoutGen      int64
}

func NewCredentialTargetResolverBinder() (platformtx.CredentialResolverBinder, error) {
	return platformtx.NewCredentialResolverBinder(
		func(database platformpostgres.BoundExecutor) platformtx.CredentialTargetResolver {
			return &credentialTargetResolver{database: database}
		},
	)
}

func (resolver *credentialTargetResolver) CredentialTarget(
	ctx context.Context,
	lookup platformtx.CredentialLookup,
) (platformtx.CredentialTargetEvidence, error) {
	if resolver == nil || ctx == nil {
		return platformtx.CredentialTargetEvidence{}, errCredentialTargetUnavailable
	}
	var row credentialTargetRow
	var err error
	switch lookup.LocatorKind() {
	case platformtx.CredentialLoginDirectory:
		row, err = resolver.queryLoginTarget(ctx,
			lookup.TenantID(), string(lookup.IdentifierKind()), lookup.KeyID(), lookup.Key(),
		)
	case platformtx.CredentialRefreshLocator:
		row, err = resolver.queryRefreshTarget(ctx, lookup.Key())
	case platformtx.CredentialTenantRoute, platformtx.CredentialSessionLocator:
		row, err = resolver.queryTenantTarget(ctx, lookup.TenantID(), lookup.Key())
	default:
		return platformtx.CredentialTargetEvidence{}, errCredentialTargetUnavailable
	}
	if err != nil {
		return platformtx.CredentialTargetEvidence{}, err
	}
	if lookup.LocatorKind() == platformtx.CredentialLoginDirectory &&
		row.tenantID != lookup.TenantID() {
		return platformtx.CredentialTargetEvidence{}, errCredentialTargetUnavailable
	}
	var binding [32]byte
	copy(binding[:], row.sourceSHA256)
	return platformtx.CredentialTargetEvidence{
		OperationID: lookup.OperationID(), TenantID: row.tenantID,
		Purpose: lookup.Purpose(), LocatorKind: lookup.LocatorKind(),
		SourceUserID: row.userID, SourceIdentifierID: row.identifierID,
		SourceIdentifierKind: row.identifierKind,
		SourceTokenID:        row.tokenID, SourceFamilyID: row.familyID,
		SourceRequestID: row.requestID, SourceGrantID: row.grantID,
		SourceSelector:       row.selector,
		LocatorBindingSHA256: binding, SourceRevision: row.sourceRevision,
		GroupGeneration: row.groupGen, IAMGeneration: row.iamGen,
		SystemResourceGeneration: row.layoutGen,
	}, nil
}
