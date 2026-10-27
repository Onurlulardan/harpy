package iam

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *permissionRequestHTTPApplication) Cancel(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestTerminal,
) (PermissionRequestRefView, error) {
	if input.ID == uuid.Nil || input.ExpectedRevision < 1 ||
		!permissionrequest.CanTransition(input.ExpectedState, permissionrequest.StateCancelled) {
		return PermissionRequestRefView{}, ErrPermissionRequestHTTPInvalid
	}
	operationID := permissionRequestTerminalOperationID(claims, "cancel", input)
	requestSHA, err := permissionRequestRequestSHA(input)
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	command := permissionrequest.CancelCommand{
		OperationID: operationID, RequestSHA256: requestSHA,
		GroupID: input.ID, ExpectedRevision: input.ExpectedRevision,
		ExpectedState: input.ExpectedState, RequesterID: claims.Subject(),
		ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
	}
	var result permissionrequest.GroupRef
	err = application.inCommand(
		ctx, claims, operationID.String(), "",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var executeErr error
			result, executeErr = application.runtime.service.Cancel(scoped, scope, command)
			return executeErr
		},
	)
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	return permissionRequestRefView(claims.Tenant(), result), nil
}

func (application *permissionRequestHTTPApplication) Expire(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestTerminal,
) (PermissionRequestRefView, error) {
	if input.ID == uuid.Nil || input.ExpectedRevision < 1 ||
		!permissionrequest.CanTransition(input.ExpectedState, permissionrequest.StateExpired) {
		return PermissionRequestRefView{}, ErrPermissionRequestHTTPInvalid
	}
	operationID := permissionRequestTerminalOperationID(claims, "expire", input)
	requestSHA, err := permissionRequestRequestSHA(input)
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	command := permissionrequest.ExpireCommand{
		OperationID: operationID, RequestSHA256: requestSHA,
		GroupID: input.ID, ExpectedRevision: input.ExpectedRevision,
		ExpectedState: input.ExpectedState,
	}
	var result permissionrequest.GroupRef
	err = application.inCommand(
		ctx, claims, operationID.String(), "manage",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var executeErr error
			result, executeErr = application.runtime.service.Expire(scoped, scope, command)
			return executeErr
		},
	)
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	return permissionRequestRefView(claims.Tenant(), result), nil
}

func permissionRequestTerminalOperationID(
	claims accesstoken.Claims,
	operation string,
	input PermissionRequestTerminal,
) uuid.UUID {
	key := fmt.Sprintf("%s:%d:%s", input.ID, input.ExpectedRevision, input.ExpectedState)
	return permissionRequestOperationID(
		claims, "iam.permission_requests."+operation, key,
	)
}
