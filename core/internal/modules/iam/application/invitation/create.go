package invitation

import (
	"context"
	"unicode/utf8"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type CreateCommand struct {
	OperationID                  uuid.UUID
	RequestSHA256                [32]byte
	ID                           uuid.UUID
	DisplayValue                 string
	NormalizedValue              string
	NormalizedSHA256             [32]byte
	RoleID                       uuid.UUID
	RoleScope                    RoleScope
	InvitedBy                    uuid.UUID
	ExpectedAuthorizationVersion int64
	Token                        TokenIssue
}

func (command CreateCommand) Validate() error {
	if command.OperationID == uuid.Nil || command.RequestSHA256 == [32]byte{} ||
		command.ID == uuid.Nil || command.DisplayValue == "" ||
		utf8.RuneCountInString(command.DisplayValue) > 320 ||
		len(command.NormalizedValue) < 3 || len(command.NormalizedValue) > 254 ||
		command.NormalizedSHA256 == [32]byte{} || command.RoleID == uuid.Nil ||
		command.RoleScope.Validate() != nil || command.InvitedBy == uuid.Nil ||
		command.ExpectedAuthorizationVersion < 1 || command.Token.Validate() != nil ||
		command.Token.DeliveryEnvelope.InvitationID != command.ID {
		return ErrInvalid
	}
	return nil
}

func (service *Service) Create(
	ctx context.Context,
	scope platformtx.TenantScope,
	command CreateCommand,
) (Ref, error) {
	if service == nil || service.store == nil || service.normalizer == nil ||
		service.effects == nil ||
		command.Validate() != nil || command.Token.MatchesTenant(scope.TenantID()) != nil ||
		scope.ActorID() != command.InvitedBy {
		return Ref{}, ErrInvalid
	}
	evidence, err := service.normalizer.NormalizeEmail(command.DisplayValue)
	if err != nil || evidence.NormalizedValue != command.NormalizedValue ||
		evidence.SHA256 != command.NormalizedSHA256 {
		return Ref{}, ErrInvalid
	}
	result, err := service.store.Create(ctx, scope, command)
	if err != nil {
		return Ref{}, err
	}
	if result.Validate() != nil || result.State != StatePending {
		return Ref{}, ErrUnavailable
	}
	if result.Replayed {
		return result, nil
	}
	if err := service.effects.WriteTenantChange(ctx, scope, AuditChange{
		OperationID: command.OperationID, RequestSHA256: command.RequestSHA256,
		InvitationID: command.ID, ActorID: command.InvitedBy,
		CommandKey: "iam.invitation.create", ResultState: result.State,
		ResultRevision:      result.Revision,
		DeliveryArtifactRef: command.Token.DeliveryArtifactRef,
	}); err != nil {
		return Ref{}, ErrEffects
	}
	return result, nil
}
