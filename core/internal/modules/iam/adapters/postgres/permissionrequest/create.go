package postgres

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

type permissionRequestCreateDocument struct {
	ID               uuid.UUID `json:"id"`
	Ordinal          int16     `json:"ordinal"`
	TargetKind       string    `json:"target_kind"`
	TargetID         uuid.UUID `json:"target_id"`
	PermissionID     uuid.UUID `json:"permission_id"`
	PlannedBindingID uuid.UUID `json:"planned_binding_id"`
}

func (store *permissionRequestStore) Create(
	ctx context.Context,
	scope tenantScope,
	command permissionrequest.CreateCommand,
) (permissionrequest.GroupRef, error) {
	if err := command.Validate(); err != nil ||
		!permissionRequestCommandScopeUsable(ctx, scope) ||
		scope.ActorID() != command.RequesterID {
		return permissionrequest.GroupRef{}, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return permissionrequest.GroupRef{}, err
	}
	authority, err := lookupPermissionRequestAuthority(ctx, bound)
	if err != nil {
		return permissionrequest.GroupRef{}, err
	}
	documents := make([]permissionRequestCreateDocument, 0, len(command.Requests))
	for _, item := range command.Requests {
		documents = append(documents, permissionRequestCreateDocument{
			ID: item.ID, Ordinal: item.Ordinal, TargetKind: string(item.Target.Kind),
			TargetID: item.Target.ID, PermissionID: item.PermissionID,
			PlannedBindingID: item.PlannedBindingID,
		})
	}
	payload, err := json.Marshal(documents)
	if err != nil {
		return permissionrequest.GroupRef{}, permissionrequest.ErrInvalid
	}
	return scanPermissionRequestReference(bound.database.QueryRow(
		ctx, permissionRequestCreateSQL,
		command.GroupID, command.RequesterID, string(command.Mode),
		authority.manage, authority.review,
		command.ExpectedAuthorizationVersion,
		command.StepUpPurposeSHA256[:], command.ExpiresAt, payload,
	))
}

type permissionRequestAuthority struct {
	manage uuid.UUID
	review uuid.UUID
}

func lookupPermissionRequestAuthority(
	ctx context.Context,
	bound *boundPermissionRequestStore,
) (permissionRequestAuthority, error) {
	var authority permissionRequestAuthority
	if err := bound.database.QueryRow(ctx, permissionRequestAuthoritySQL).Scan(
		&authority.manage, &authority.review,
	); err != nil || authority.manage == uuid.Nil || authority.review == uuid.Nil ||
		authority.manage == authority.review {
		return permissionRequestAuthority{}, permissionrequest.ErrUnavailable
	}
	return authority, nil
}
