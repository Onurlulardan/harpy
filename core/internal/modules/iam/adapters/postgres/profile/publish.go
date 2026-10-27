package profile

import (
	"context"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func (store *commandStore) Publish(
	ctx context.Context,
	scope tenantScope,
	command profilecommand.PublishCommand,
) (profilecommand.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != command.Evidence.ActorID {
		return profilecommand.MutationResult{}, profilecommand.ErrInvalidCommand
	}
	bound, err := store.bind(scope)
	if err != nil {
		return profilecommand.MutationResult{}, err
	}
	change, evidence := command.Change, command.Evidence
	schema := change.Schema
	hash := schema.DefinitionSHA256()
	return scanMutation(bound.database.QueryRow(
		ctx, publishSQL, evidence.OperationID, evidence.RequestSHA256[:],
		schema.ID(), string(schema.Resource()), schema.Version(),
		string(schema.DefinitionJSON()), hash[:], change.ExpectedActiveSchemaID,
		evidence.ActorID, evidence.ManagePermissionID,
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}
