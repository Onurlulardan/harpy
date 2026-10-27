package portaluser

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) Create(
	ctx context.Context, scope platformtx.TenantScope, command CreateCommand,
) (MutationResult, error) {
	if service == nil || !scopeUsable(ctx, scope, command.Evidence, platformtx.Command) ||
		command.UserID == [16]byte{} || command.Profile.Validate(identitydomain.ProfileUser) != nil ||
		!command.Portal.Valid() || !command.Email.Valid() || !command.Subject.Valid() ||
		(command.SetupState != "PENDING" && command.SetupState != "INACTIVE") ||
		len(command.RoleNames) == 0 || len(command.RoleNames) != len(command.RoleAssignmentIDs) {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.Create(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !result.Valid() {
		return MutationResult{}, ErrUnavailable
	}
	return result, nil
}
