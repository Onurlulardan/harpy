package invitation

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type AcceptCommand struct {
	OperationID                  uuid.UUID
	RequestSHA256                [32]byte
	ID                           uuid.UUID
	ExpectedRevision             int64
	Token                        TokenProof
	AcceptedUserID               uuid.UUID
	AcceptedIdentifierID         uuid.UUID
	AccessModeID                 uuid.UUID
	RoleAssignmentID             uuid.UUID
	OrganizationMembershipID     uuid.UUID
	Profile                      RegistrationProfile
	Password                     RegistrationPassword
	Directory                    DirectoryProjection
	ExpectedAuthorizationVersion int64
}

type AcceptResult struct {
	Code               string
	Revision           int64
	AuthorizationClock int64
	Replayed           bool
}

func (command AcceptCommand) Validate() error {
	if command.OperationID == uuid.Nil || command.RequestSHA256 == [32]byte{} ||
		command.ID == uuid.Nil || command.ExpectedRevision < 1 ||
		command.Token.Validate() != nil || command.AcceptedUserID == uuid.Nil ||
		command.AcceptedIdentifierID == uuid.Nil ||
		command.AccessModeID == uuid.Nil ||
		command.RoleAssignmentID == uuid.Nil ||
		command.Profile.Validate() != nil || command.Password.Validate() != nil ||
		command.Directory.Validate() != nil ||
		command.ExpectedAuthorizationVersion < 1 {
		return ErrInvalid
	}
	return nil
}

func (service *Service) Accept(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	command AcceptCommand,
) (AcceptResult, error) {
	if service == nil || service.store == nil || service.effects == nil ||
		command.Validate() != nil ||
		platformtx.RequireCredentialPurpose(
			scope, platformtx.CredentialInvitation,
		) != nil || scope.LocatorBindingSHA256() != command.Token.BindingSHA256 {
		return AcceptResult{}, ErrInvalid
	}
	result, err := service.store.Accept(ctx, scope, command)
	if err != nil {
		return AcceptResult{}, err
	}
	if result.Code != "iam.invitation.accepted" || result.Revision < 2 ||
		result.AuthorizationClock < 1 {
		return AcceptResult{}, ErrUnavailable
	}
	if result.Replayed {
		return result, nil
	}
	if err := service.effects.WriteCredentialChange(ctx, scope, AuditChange{
		OperationID: command.OperationID, InvitationID: command.ID,
		RequestSHA256: command.RequestSHA256, ActorID: command.AcceptedUserID,
		CommandKey: "iam.invitation.accept", ResultState: StateAccepted,
		ResultRevision:     result.Revision,
		AuthorizationClock: result.AuthorizationClock,
	}); err != nil {
		return AcceptResult{}, ErrEffects
	}
	return result, nil
}
