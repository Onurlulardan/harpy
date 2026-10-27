package portaluserpostgres

import (
	"context"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

func (storage *store) Create(
	ctx context.Context, scope tenantScope, command portaluser.CreateCommand,
) (portaluser.MutationResult, error) {
	if !usable(ctx, scope, commandMode) || !command.Evidence.Valid() ||
		scope.ActorID() != command.Evidence.ActorID ||
		command.Profile.Validate(identitydomain.ProfileUser) != nil {
		return portaluser.MutationResult{}, portaluser.ErrInvalid
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return portaluser.MutationResult{}, err
	}
	profileSHA := command.Profile.SHA256()
	evidence := command.Evidence
	result, err := scanMutation(bound.database.QueryRow(ctx, createSQL,
		scope.TenantID(), evidence.OperationID, evidence.RequestSHA256[:], command.UserID,
		command.Profile.SchemaID(), command.Profile.Values(), profileSHA[:],
		command.Portal.ID, command.Portal.Display, command.Portal.Normalized,
		command.Portal.NormalizedSHA256[:], command.Email.ID, command.Email.Display,
		command.Email.Normalized, command.Email.NormalizedSHA256[:],
		command.Subject.AccessModeID, command.Subject.BindingID, command.Subject.Type,
		command.Subject.FormID, command.Subject.FormVersionID, command.Subject.RecordID,
		command.Subject.IdentifierFieldID, command.Subject.ContractSHA256[:],
		command.SetupState, command.RoleNames, command.RoleAssignmentIDs,
		evidence.ActorID, evidence.PermissionID, evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	))
	if err != nil {
		return portaluser.MutationResult{}, err
	}
	if err = storage.projectDirectory(ctx, scope, bound, command.Email,
		result.ID, result.EmailIdentifierID, result.EmailRevision, "email_v1"); err != nil {
		return portaluser.MutationResult{}, err
	}
	if err = storage.projectDirectory(ctx, scope, bound, command.Portal,
		result.ID, result.PortalIdentifierID, result.PortalRevision, "portal_v1"); err != nil {
		return portaluser.MutationResult{}, err
	}
	return result, nil
}

func (storage *store) projectDirectory(
	ctx context.Context, scope tenantScope, bound *boundStore,
	identifier portaluser.Identifier, userID, identifierID uuid.UUID,
	revision int64, kind string,
) error {
	authority, err := mintIAMOwnerAuthority(scope)
	if err != nil || !authority.OwnsScope(scope) || authority.GroupID() != "iam_system_resources" {
		return portaluser.ErrUnavailable
	}
	var projected int64
	err = bound.database.QueryRow(ctx, directoryUpsertSQL,
		scope.TenantID(), authority.GroupGeneration(), authority.Generation(),
		userID, identifierID, kind, identifier.LookupKeyID,
		identifier.LookupHMAC[:], revision, identifier.NormalizedSHA256[:],
	).Scan(&projected)
	if err != nil || projected != revision {
		return portaluser.ErrUnavailable
	}
	return nil
}
