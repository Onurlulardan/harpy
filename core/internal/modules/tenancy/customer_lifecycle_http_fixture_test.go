package tenancy

import (
	"context"
	"testing"
	"time"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/google/uuid"
)

var (
	httpLifecycleTenantID    = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	httpLifecycleOperationID = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	httpLifecycleNow         = time.Date(2026, 8, 2, 3, 23, 0, 0, time.UTC)
)

type lifecycleHTTPFixture struct {
	listQuery       lifecycle.ListQuery
	getQuery        lifecycle.GetQuery
	mutation        lifecycle.MutationCommand
	async           lifecycle.AsyncCommand
	operationQuery  lifecycle.OperationQuery
	listResult      lifecycle.Page
	tenantResult    lifecycle.Tenant
	operationResult lifecycle.Operation
	err             error
	lastCall        string
}

func (fixture *lifecycleHTTPFixture) List(_ context.Context, query lifecycle.ListQuery) (lifecycle.Page, error) {
	fixture.lastCall, fixture.listQuery = "list", query
	return fixture.listResult, fixture.err
}
func (fixture *lifecycleHTTPFixture) Get(_ context.Context, query lifecycle.GetQuery) (lifecycle.Tenant, error) {
	fixture.lastCall, fixture.getQuery = "get", query
	return fixture.tenantResult, fixture.err
}
func (fixture *lifecycleHTTPFixture) Update(_ context.Context, command lifecycle.MutationCommand) (lifecycle.Tenant, error) {
	fixture.lastCall, fixture.mutation = "update", command
	return fixture.tenantResult, fixture.err
}
func (fixture *lifecycleHTTPFixture) Suspend(_ context.Context, command lifecycle.MutationCommand) (lifecycle.Tenant, error) {
	fixture.lastCall, fixture.mutation = "suspend", command
	return fixture.tenantResult, fixture.err
}
func (fixture *lifecycleHTTPFixture) Reactivate(_ context.Context, command lifecycle.MutationCommand) (lifecycle.Tenant, error) {
	fixture.lastCall, fixture.mutation = "reactivate", command
	return fixture.tenantResult, fixture.err
}
func (fixture *lifecycleHTTPFixture) RequestOffboard(_ context.Context, command lifecycle.AsyncCommand) (lifecycle.Operation, error) {
	fixture.lastCall, fixture.async = "offboard", command
	return fixture.operationResult, fixture.err
}
func (fixture *lifecycleHTTPFixture) RequestPurge(_ context.Context, command lifecycle.AsyncCommand) (lifecycle.Operation, error) {
	fixture.lastCall, fixture.async = "purge", command
	return fixture.operationResult, fixture.err
}
func (fixture *lifecycleHTTPFixture) ReadOperation(_ context.Context, query lifecycle.OperationQuery) (lifecycle.Operation, error) {
	fixture.lastCall, fixture.operationQuery = "read_operation", query
	return fixture.operationResult, fixture.err
}

func newLifecycleHTTPFixture(t *testing.T) (*CustomerLifecycleHTTP, *lifecycleHTTPFixture) {
	t.Helper()
	fixture := &lifecycleHTTPFixture{
		listResult: lifecycle.Page{Items: []lifecycle.Tenant{}},
		tenantResult: lifecycle.Tenant{ID: httpLifecycleTenantID, Name: "Tenant",
			Slug: "tenant", Status: "ACTIVE", Revision: 2, SettingsRevision: 1,
			CreatedAt: httpLifecycleNow, UpdatedAt: httpLifecycleNow},
		operationResult: lifecycle.Operation{ID: httpLifecycleOperationID,
			TargetTenantID: httpLifecycleTenantID, Kind: lifecycle.OperationOffboard,
			State: "REQUESTED", Revision: 2, CreatedAt: httpLifecycleNow,
			UpdatedAt: httpLifecycleNow},
	}
	handler, err := NewCustomerLifecycleHTTP(fixture, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	return handler, fixture
}
