package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

type permissionRequestHTTPTestApplication struct {
	createInput   PermissionRequestCreate
	decisionInput PermissionRequestDecision
	createResult  PermissionRequestRefView
	pageResult    PermissionRequestPageView
	resourcePage  PermissionRequestCatalogResourcePageView
	actionPage    PermissionRequestCatalogActionPageView
	result        PermissionRequestRefView
	err           error
	createCalls   int
	listCalls     int
	resourceCalls int
	actionCalls   int
	missingCalls  int
	decisionCalls int
}

func (app *permissionRequestHTTPTestApplication) ListCatalogResources(
	_ context.Context, _ accesstoken.Claims, _ PermissionRequestCatalogList,
) (PermissionRequestCatalogResourcePageView, error) {
	app.resourceCalls++
	return app.resourcePage, app.err
}

func (app *permissionRequestHTTPTestApplication) ListCatalogActions(
	_ context.Context, _ accesstoken.Claims, _ PermissionRequestCatalogList,
) (PermissionRequestCatalogActionPageView, error) {
	app.actionCalls++
	return app.actionPage, app.err
}

func (app *permissionRequestHTTPTestApplication) ListMissingCatalogResources(
	_ context.Context, _ accesstoken.Claims, _ PermissionRequestMissingCatalogList,
) (PermissionRequestCatalogResourcePageView, error) {
	app.missingCalls++
	return app.resourcePage, app.err
}

func (app *permissionRequestHTTPTestApplication) Create(
	_ context.Context, _ accesstoken.Claims, input PermissionRequestCreate,
) (PermissionRequestRefView, error) {
	app.createCalls++
	app.createInput = input
	return app.createResult, app.err
}

func (app *permissionRequestHTTPTestApplication) List(
	_ context.Context, _ accesstoken.Claims, _ PermissionRequestList,
) (PermissionRequestPageView, error) {
	app.listCalls++
	return app.pageResult, app.err
}

func (app *permissionRequestHTTPTestApplication) Get(
	context.Context, accesstoken.Claims, uuid.UUID,
) (PermissionRequestGroupView, error) {
	return PermissionRequestGroupView{}, app.err
}

func (app *permissionRequestHTTPTestApplication) Review(
	context.Context, accesstoken.Claims, PermissionRequestReview,
) (PermissionRequestRefView, error) {
	return app.result, app.err
}

func (app *permissionRequestHTTPTestApplication) Approve(
	_ context.Context, _ accesstoken.Claims, input PermissionRequestDecision,
) (PermissionRequestRefView, error) {
	app.decisionCalls++
	app.decisionInput = input
	return app.result, app.err
}

func (app *permissionRequestHTTPTestApplication) Reject(
	ctx context.Context, claims accesstoken.Claims, input PermissionRequestDecision,
) (PermissionRequestRefView, error) {
	return app.Approve(ctx, claims, input)
}

func (app *permissionRequestHTTPTestApplication) Cancel(
	context.Context, accesstoken.Claims, PermissionRequestTerminal,
) (PermissionRequestRefView, error) {
	return app.result, app.err
}

func (app *permissionRequestHTTPTestApplication) Expire(
	context.Context, accesstoken.Claims, PermissionRequestTerminal,
) (PermissionRequestRefView, error) {
	return app.result, app.err
}
