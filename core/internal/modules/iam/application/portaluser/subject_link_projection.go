package portaluser

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const MaximumSubjectLinkBatch = 500

func (service *Service) ProjectSubjectLinks(
	ctx context.Context, scope platformtx.TenantScope,
	access ProjectionAccess, formID, versionID uuid.UUID, recordIDs []uuid.UUID,
) ([]SubjectLink, error) {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!access.Valid() || formID == uuid.Nil || versionID == uuid.Nil || len(recordIDs) == 0 ||
		len(recordIDs) > MaximumSubjectLinkBatch || !uniqueUUIDs(recordIDs) ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil ||
		scope.ActorID() != access.PrincipalID || scope.TenantID() != access.TenantID {
		return nil, ErrInvalid
	}
	result, err := service.store.ProjectSubjectLinks(
		ctx, scope, access, formID, versionID, recordIDs,
	)
	if err != nil {
		return nil, err
	}
	seen := make(map[uuid.UUID]struct{}, len(result))
	for _, link := range result {
		if !link.Valid() || !containsUUID(recordIDs, link.SubjectRecordID) {
			return nil, ErrUnavailable
		}
		if _, duplicate := seen[link.SubjectRecordID]; duplicate {
			return nil, ErrUnavailable
		}
		seen[link.SubjectRecordID] = struct{}{}
	}
	return result, nil
}

func uniqueUUIDs(values []uuid.UUID) bool {
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func containsUUID(values []uuid.UUID, wanted uuid.UUID) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
