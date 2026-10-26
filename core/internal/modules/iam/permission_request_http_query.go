package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *permissionRequestHTTPApplication) Get(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (PermissionRequestGroupView, error) {
	if id == uuid.Nil {
		return PermissionRequestGroupView{}, ErrPermissionRequestHTTPInvalid
	}
	var result permissionrequest.Group
	err := application.inSnapshot(
		ctx, claims,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var err error
			result, err = application.runtime.service.Read(
				scoped, scope, permissionrequest.ReadQuery{GroupID: id},
			)
			return err
		},
	)
	if err != nil {
		return PermissionRequestGroupView{}, err
	}
	return permissionRequestGroupView(claims.Tenant(), result), nil
}

func (application *permissionRequestHTTPApplication) List(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestList,
) (PermissionRequestPageView, error) {
	return application.listPermissionRequestsWithEvidence(ctx, claims, input, nil)
}

func (application *permissionRequestHTTPApplication) listPermissionRequestsWithEvidence(
	ctx context.Context,
	evidence identityEvidenceSource,
	input PermissionRequestList,
	actions []string,
) (PermissionRequestPageView, error) {
	query := permissionrequest.ListQuery{Limit: input.Limit, After: input.After}
	if query.Validate() != nil {
		return PermissionRequestPageView{}, ErrPermissionRequestHTTPInvalid
	}
	var result permissionrequest.Page
	execute := func(scoped context.Context, scope platformtx.TenantScope) error {
		var err error
		result, err = application.runtime.service.List(scoped, scope, query)
		return err
	}
	var err error
	if len(actions) == 0 {
		err = application.inSnapshot(ctx, evidence, execute)
	} else {
		err = application.inSnapshotAnyAction(ctx, evidence, actions, execute)
	}
	if err != nil {
		return PermissionRequestPageView{}, err
	}
	view := PermissionRequestPageView{
		Items:   make([]PermissionRequestGroupView, len(result.Groups)),
		HasMore: result.Next != uuid.Nil,
	}
	for index, group := range result.Groups {
		view.Items[index] = permissionRequestGroupView(evidence.Tenant(), group)
	}
	if result.Next != uuid.Nil {
		value := result.Next.String()
		view.NextAfterID = &value
	}
	return view, nil
}
