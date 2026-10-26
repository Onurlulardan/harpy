package passwordenrollmentpostgres

import (
	"context"
	_ "embed"

	passwordenrollment "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/passwordenrollment"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

//go:embed sql/tenant/enroll.sql
var enrollSQL string

func (storage *store) Create(
	ctx context.Context,
	scope passwordenrollment.TenantScope,
	command passwordenrollment.Command,
) (passwordenrollment.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope, commandMode) ||
		scope.ActorID() != uuid.UUID(command.Evidence.ActorID) {
		return passwordenrollment.MutationResult{}, passwordenrollment.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return passwordenrollment.MutationResult{}, err
	}
	salt, tag := command.Credential.Salt(), command.Credential.Tag()
	evidence := command.Evidence
	var result identitydomain.MutationResult
	var resultID uuid.UUID
	err = bound.database.QueryRow(
		ctx, enrollSQL, databaseID(evidence.OperationID), evidence.RequestSHA256[:],
		databaseID(command.UserID), salt[:], tag[:], databaseID(evidence.ActorID),
		databaseID(evidence.PermissionID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	).Scan(&resultID, &result.Revision, &result.AuthorizationClock, &result.Replayed)
	if err != nil {
		return passwordenrollment.MutationResult{}, passwordenrollment.ErrStorage
	}
	result.ID = identitydomain.IdentityID(resultID)
	if result.Validate() != nil {
		return passwordenrollment.MutationResult{}, passwordenrollment.ErrStorage
	}
	return result, nil
}
