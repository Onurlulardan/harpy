package profile

import (
	"context"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func (store *commandStore) Retire(
	ctx context.Context,
	scope tenantScope,
	command profilecommand.RetireCommand,
) (profilecommand.MutationResult, error) {
	if command.Validate() != nil || command.References.Validate(command.Target) != nil ||
		!scopeUsable(ctx, scope) || scope.ActorID() != command.Evidence.ActorID {
		return profilecommand.MutationResult{}, profilecommand.ErrInvalidCommand
	}
	bound, err := store.bind(scope)
	if err != nil {
		return profilecommand.MutationResult{}, err
	}
	change, evidence, references := command.Change, command.Evidence, command.References
	schema := change.Schema
	hash := schema.DefinitionSHA256()
	return scanMutation(bound.database.QueryRow(
		ctx, retireSQL, evidence.OperationID, evidence.RequestSHA256[:],
		schema.ID(), string(schema.Resource()), schema.Version(),
		string(schema.DefinitionJSON()), hash[:], change.ExpectedActiveSchemaID,
		command.Target.FieldID, command.Target.ExpectedFieldRevision,
		references.ManifestSHA256[:], references.CheckedOwners,
		evidence.ActorID, evidence.ManagePermissionID,
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}
