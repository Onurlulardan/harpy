package profile

import (
	"context"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func (store *commandStore) Backfill(
	ctx context.Context,
	scope tenantScope,
	command profilecommand.BackfillCommand,
) (profilecommand.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != command.Evidence.ActorID {
		return profilecommand.MutationResult{}, profilecommand.ErrInvalidCommand
	}
	bound, err := store.bind(scope)
	if err != nil {
		return profilecommand.MutationResult{}, err
	}
	evidence, document := command.Evidence, command.Target
	hash := document.SHA256()
	return scanMutation(bound.database.QueryRow(
		ctx, backfillSQL, evidence.OperationID, evidence.RequestSHA256[:],
		string(document.Resource()), command.ResourceID,
		command.ExpectedSourceSchemaID, document.SchemaID(),
		document.CanonicalJSON(), hash[:], command.ExpectedRevision,
		evidence.ActorID, evidence.ManagePermissionID,
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}
