package iam

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *permissionRequestHTTPApplication) Create(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestCreate,
) (PermissionRequestRefView, error) {
	if !permissionRequestCreateValid(input) {
		return PermissionRequestRefView{}, ErrPermissionRequestHTTPInvalid
	}
	operationID := permissionRequestOperationID(
		claims, "iam.permission_requests.create", input.IdempotencyKey,
	)
	groupID := permissionRequestChildID(operationID, "group")
	requestSHA, err := permissionRequestRequestSHA(input)
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	command := permissionrequest.CreateCommand{
		OperationID: operationID, RequestSHA256: requestSHA,
		GroupID: groupID, RequesterID: claims.Subject(), Mode: input.Mode,
		ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
		StepUpPurposeSHA256:          permissionRequestPurpose(claims.Tenant(), groupID),
		ExpiresAt:                    input.ExpiresAt,
		Requests:                     make([]permissionrequest.CreateItem, len(input.Items)),
	}
	for index, item := range input.Items {
		command.Requests[index] = permissionrequest.CreateItem{
			ID:           permissionRequestChildID(operationID, fmt.Sprintf("request/%d", index+1)),
			Ordinal:      int16(index + 1),
			Target:       permissionrequest.Target{Kind: item.TargetKind, ID: item.TargetID},
			PermissionID: item.PermissionID,
			PlannedBindingID: permissionRequestChildID(
				operationID, fmt.Sprintf("binding/%d", index+1),
			),
		}
	}
	action, _ := permissionRequestCreateAction(input.Mode)
	var result permissionrequest.GroupRef
	err = application.inCommand(
		ctx, claims, operationID.String(), action,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var err error
			result, err = application.runtime.service.Create(scoped, scope, command)
			return err
		},
	)
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	return permissionRequestRefView(claims.Tenant(), result), nil
}

func permissionRequestCreateAction(mode permissionrequest.Mode) (string, bool) {
	switch mode {
	case permissionrequest.ModeSelfService:
		return "", true
	case permissionrequest.ModePrivileged:
		return "manage", true
	default:
		return "", false
	}
}

func permissionRequestCreateValid(input PermissionRequestCreate) bool {
	if _, modeValid := permissionRequestCreateAction(input.Mode); !modeValid || input.ExpiresAt.IsZero() ||
		len(input.Items) < 1 || len(input.Items) > 100 ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return false
	}
	for _, item := range input.Items {
		if item.TargetID == uuid.Nil || item.PermissionID == uuid.Nil ||
			(permissionrequest.Target{Kind: item.TargetKind, ID: item.TargetID}).Validate() != nil {
			return false
		}
	}
	return true
}
