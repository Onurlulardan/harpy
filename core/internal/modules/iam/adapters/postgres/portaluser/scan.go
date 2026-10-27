package portaluserpostgres

import (
	"time"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	"github.com/google/uuid"
)

type scanner interface{ Scan(...any) error }

func scanMutation(row scanner) (portaluser.MutationResult, error) {
	var result portaluser.MutationResult
	if row == nil {
		return result, portaluser.ErrUnavailable
	}
	err := row.Scan(
		&result.ID, &result.Revision, &result.AuthorizationClock, &result.Replayed,
		&result.EmailIdentifierID, &result.EmailRevision,
		&result.PortalIdentifierID, &result.PortalRevision,
	)
	if err != nil {
		return portaluser.MutationResult{}, mapError(err)
	}
	if !result.Valid() || result.EmailIdentifierID == uuid.Nil ||
		result.PortalIdentifierID == uuid.Nil || result.EmailRevision < 1 || result.PortalRevision < 1 {
		return portaluser.MutationResult{}, portaluser.ErrUnavailable
	}
	return result, nil
}

func scanView(row scanner) (portaluser.View, error) {
	var result portaluser.View
	if row == nil {
		return result, portaluser.ErrUnavailable
	}
	var createdAt, updatedAt time.Time
	err := row.Scan(
		&result.ID, &result.BindingID, &result.Email, &result.Status,
		&result.AccountType, &result.SubjectType, &result.SubjectRecordID,
		&result.Revision, &result.SetupRequired, &createdAt, &updatedAt,
	)
	if err != nil {
		return portaluser.View{}, mapError(err)
	}
	result.CreatedAt, result.UpdatedAt = createdAt.UTC(), updatedAt.UTC()
	if !result.Valid() {
		return portaluser.View{}, portaluser.ErrUnavailable
	}
	return result, nil
}

func scanAccess(row scanner) (portaluser.AccessBinding, error) {
	var result portaluser.AccessBinding
	if row == nil {
		return result, portaluser.ErrUnavailable
	}
	err := row.Scan(
		&result.PrincipalID, &result.BindingID, &result.TenantID,
		&result.TokenVersion, &result.AuthorizationVersion,
		&result.AuthorizationClock, &result.BindingRevision, &result.SubjectType,
		&result.FormID, &result.FormVersionID, &result.SubjectRecordID,
		&result.IdentifierFieldID, &result.RoleNames,
	)
	if err != nil {
		return portaluser.AccessBinding{}, mapError(err)
	}
	if !result.Valid() {
		return portaluser.AccessBinding{}, portaluser.ErrUnavailable
	}
	return result, nil
}
