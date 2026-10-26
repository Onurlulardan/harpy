package portaluserpostgres

import (
	"context"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	"github.com/google/uuid"
)

func (storage *store) ProjectSubjectLinks(
	ctx context.Context, scope tenantScope,
	access portaluser.ProjectionAccess,
	formID, versionID uuid.UUID, recordIDs []uuid.UUID,
) ([]portaluser.SubjectLink, error) {
	if !access.Valid() || formID == uuid.Nil || versionID == uuid.Nil || len(recordIDs) == 0 ||
		!usable(ctx, scope, snapshotMode) {
		return nil, portaluser.ErrInvalid
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return nil, err
	}
	rows, err := bound.database.Query(
		ctx, projectSubjectLinksSQL, access.TenantID, access.PrincipalID,
		access.TokenVersion, access.AuthorizationVersion, access.AuthorizationClock,
		formID, versionID, recordIDs,
	)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	result := make([]portaluser.SubjectLink, 0, len(recordIDs))
	for rows.Next() {
		var link portaluser.SubjectLink
		if err = rows.Scan(&link.SubjectRecordID, &link.UserID, &link.Status,
			&link.AccountType, &link.Revision); err != nil || !link.Valid() {
			return nil, portaluser.ErrUnavailable
		}
		result = append(result, link)
	}
	if rows.Err() != nil {
		return nil, portaluser.ErrUnavailable
	}
	return result, nil
}
