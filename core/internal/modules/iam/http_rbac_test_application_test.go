package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

type rbacHTTPTestApplication struct {
	calls []string
	last  any
}

func (app *rbacHTTPTestApplication) called(name string, input any) {
	app.calls = append(app.calls, name)
	app.last = input
}

func (app *rbacHTTPTestApplication) CreatePermission(
	_ context.Context, _ accesstoken.Claims, input RBACPermissionCreate,
) (RBACPermissionView, error) {
	app.called("CreatePermission", input)
	return testRBACPermissionView(), nil
}

func (app *rbacHTTPTestApplication) GetPermission(
	context.Context, accesstoken.Claims, uuid.UUID,
) (RBACPermissionView, error) {
	app.called("GetPermission", nil)
	return testRBACPermissionView(), nil
}

func (app *rbacHTTPTestApplication) ListPermissions(
	_ context.Context, _ accesstoken.Claims, input RBACList,
) (RBACPermissionPage, error) {
	app.called("ListPermissions", input)
	return RBACPermissionPage{Items: []RBACPermissionView{}}, nil
}

func (app *rbacHTTPTestApplication) UpdatePermission(
	_ context.Context, _ accesstoken.Claims, input RBACPermissionUpdate,
) (RBACPermissionView, error) {
	app.called("UpdatePermission", input)
	return testRBACPermissionView(), nil
}

func (app *rbacHTTPTestApplication) DeletePermission(
	_ context.Context, _ accesstoken.Claims, input RBACMutation,
) error {
	app.called("DeletePermission", input)
	return nil
}

func (app *rbacHTTPTestApplication) CreateResource(
	_ context.Context, _ accesstoken.Claims, input RBACCatalogCreate,
) (RBACResourceView, error) {
	app.called("CreateResource", input)
	return testRBACResourceView(), nil
}

func (app *rbacHTTPTestApplication) GetResource(
	context.Context, accesstoken.Claims, uuid.UUID,
) (RBACResourceView, error) {
	app.called("GetResource", nil)
	return testRBACResourceView(), nil
}

func (app *rbacHTTPTestApplication) ListResources(
	_ context.Context, _ accesstoken.Claims, input RBACList,
) (RBACResourcePage, error) {
	app.called("ListResources", input)
	return RBACResourcePage{Items: []RBACResourceView{}}, nil
}

func (app *rbacHTTPTestApplication) UpdateResource(
	_ context.Context, _ accesstoken.Claims, input RBACCatalogUpdate,
) (RBACResourceView, error) {
	app.called("UpdateResource", input)
	return testRBACResourceView(), nil
}

func (app *rbacHTTPTestApplication) DeleteResource(
	_ context.Context, _ accesstoken.Claims, input RBACMutation,
) error {
	app.called("DeleteResource", input)
	return nil
}

func (app *rbacHTTPTestApplication) CreateAction(
	_ context.Context, _ accesstoken.Claims, input RBACCatalogCreate,
) (RBACActionView, error) {
	app.called("CreateAction", input)
	return testRBACActionView(), nil
}

func (app *rbacHTTPTestApplication) GetAction(
	context.Context, accesstoken.Claims, uuid.UUID,
) (RBACActionView, error) {
	app.called("GetAction", nil)
	return testRBACActionView(), nil
}

func (app *rbacHTTPTestApplication) ListActions(
	_ context.Context, _ accesstoken.Claims, input RBACList,
) (RBACActionPage, error) {
	app.called("ListActions", input)
	return RBACActionPage{Items: []RBACActionView{}}, nil
}

func (app *rbacHTTPTestApplication) UpdateAction(
	_ context.Context, _ accesstoken.Claims, input RBACCatalogUpdate,
) (RBACActionView, error) {
	app.called("UpdateAction", input)
	return testRBACActionView(), nil
}

func (app *rbacHTTPTestApplication) DeleteAction(
	_ context.Context, _ accesstoken.Claims, input RBACMutation,
) error {
	app.called("DeleteAction", input)
	return nil
}

func (app *rbacHTTPTestApplication) Check(
	_ context.Context, _ accesstoken.Claims, input RBACCheck,
) (RBACDecisionView, error) {
	app.called("Check", input)
	return RBACDecisionView{Allowed: true, Reason: "direct_user_exact"}, nil
}

func (app *rbacHTTPTestApplication) BatchCheck(
	_ context.Context, _ accesstoken.Claims, input RBACBatchCheck,
) (RBACBatchDecisionView, error) {
	app.called("BatchCheck", input)
	return RBACBatchDecisionView{Results: map[string]bool{"iam.users:read": true}}, nil
}

func (app *rbacHTTPTestApplication) Effective(
	_ context.Context, _ accesstoken.Claims, input RBACEffectiveRead,
) (RBACEffectivePage, error) {
	app.called("Effective", input)
	return RBACEffectivePage{Items: []RBACEffectiveView{}}, nil
}
