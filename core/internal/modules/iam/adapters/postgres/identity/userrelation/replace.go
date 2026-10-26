package userrelationpostgres

import (
	"context"
	_ "embed"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/replace.sql
var replaceUserRelationsSQL string

func (storage *store) ReplaceUserRelations(
	ctx context.Context,
	scope tenantScope,
	command identitydomain.ReplaceUserRelationsCommand,
) (identitydomain.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != databaseID(command.Evidence.ActorID) {
		return identitydomain.MutationResult{}, identitydomain.ErrInvalidCommand
	}
	profileSchemaID, profileDocument, profileSHA256, valid := profileArguments(command)
	if !valid {
		return identitydomain.MutationResult{}, identitydomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identitydomain.MutationResult{}, err
	}
	evidence := command.Evidence
	return scanMutation(bound.database.QueryRow(
		ctx, replaceUserRelationsSQL,
		databaseID(evidence.OperationID), command.CommandKey, evidence.RequestSHA256[:],
		databaseID(command.UserID), command.ExpectedRevision,
		command.PatchProfile, profileSchemaID, profileDocument, profileSHA256,
		command.Organizations.Replace(), databaseIDs(command.Organizations.OrganizationIDs()),
		databaseIDs(command.Organizations.MembershipIDs()),
		command.Roles.Replace(), databaseIDs(command.Roles.RoleIDs()),
		databaseIDs(command.Roles.AssignmentIDs()),
		databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}

func profileArguments(command identitydomain.ReplaceUserRelationsCommand) (
	any,
	any,
	any,
	bool,
) {
	if !command.PatchProfile {
		return nil, nil, nil, true
	}
	profile := command.Profile()
	if profile.Validate(identitydomain.ProfileUser) != nil {
		return nil, nil, nil, false
	}
	digest := profile.SHA256()
	return databaseID(profile.SchemaID()), profile.Values(), digest[:], true
}
