package profile

import (
	"context"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func (store *commandStore) UpdateSelf(
	ctx context.Context,
	scope tenantScope,
	command profilecommand.SelfUpdateCommand,
) (profilecommand.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != command.ActorID {
		return profilecommand.MutationResult{}, profilecommand.ErrInvalidCommand
	}
	bound, err := store.bind(scope)
	if err != nil {
		return profilecommand.MutationResult{}, err
	}
	document := command.Target
	digest := document.SHA256()
	return scanMutation(bound.database.QueryRow(
		ctx, updateSelfSQL, command.OperationID, command.RequestSHA256[:],
		command.ActorID, command.ExpectedSourceSchemaID, document.SchemaID(),
		document.CanonicalJSON(), digest[:], command.ExpectedRevision,
		command.ExpectedTokenVersion, command.ExpectedAuthorizationVersion,
	))
}
