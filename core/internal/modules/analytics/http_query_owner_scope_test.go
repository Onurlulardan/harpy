package analytics

import (
	"context"
	"reflect"
	"slices"
	"testing"
	"time"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type queryOwnerGenerationFixture struct{}

func (queryOwnerGenerationFixture) ResolveTenantAnalyticsGenerations(
	context.Context,
	uuid.UUID,
) (int64, int64, error) {
	return 11, 12, nil
}

func (queryOwnerGenerationFixture) ResolveTenantBillingGenerations(
	context.Context,
	uuid.UUID,
) (int64, int64, error) {
	return 21, 22, nil
}

type queryOwnerScopeProbe struct {
	request platformtx.ScopeRequest
}

func (probe *queryOwnerScopeProbe) InTenant(
	_ context.Context,
	request platformtx.ScopeRequest,
	_ func(context.Context, platformtx.TenantScope) error,
) error {
	probe.request = request
	return nil
}

func TestDeleteUsesAnalyticsAndFormsOwnerScope(t *testing.T) {
	assertQueryOwnerScope(
		t,
		analyticshttp.PermissionQueryDelete,
		analyticshttp.OperationQueryDelete,
		func(runtime *HTTPRuntime, access analyticshttp.Access) error {
			_, err := runtime.DeleteQuery(
				context.Background(),
				access,
				queryapp.DeleteRequest{},
			)
			return err
		},
	)
}

func TestPublishUsesAnalyticsAndFormsOwnerScope(t *testing.T) {
	assertQueryOwnerScope(
		t,
		analyticshttp.PermissionQueryPublish,
		analyticshttp.OperationQueryPublish,
		func(runtime *HTTPRuntime, access analyticshttp.Access) error {
			_, err := runtime.PublishQuery(
				context.Background(),
				access,
				queryapp.PublishRequest{},
			)
			return err
		},
	)
}

func TestRevalidateUsesAnalyticsAndFormsOwnerScope(t *testing.T) {
	assertQueryOwnerScope(
		t,
		analyticshttp.PermissionQueryPublish,
		analyticshttp.OperationQueryRevalidate,
		func(runtime *HTTPRuntime, access analyticshttp.Access) error {
			_, err := runtime.RevalidateQuery(
				context.Background(),
				access,
				queryapp.PublishRequest{},
			)
			return err
		},
	)
}

func assertQueryOwnerScope(
	t *testing.T,
	permission analyticshttp.Permission,
	operation string,
	call func(*HTTPRuntime, analyticshttp.Access) error,
) {
	t.Helper()
	actorID := uuid.MustParse("019a0000-0000-7000-8000-000000000001")
	tenantID := uuid.MustParse("019a0000-0000-7000-8000-000000000002")
	access, err := analyticshttp.NewAccess(
		actorID,
		tenantID,
		1,
		1,
		1,
		time.Time{},
		permission,
		operation,
	)
	if err != nil {
		t.Fatalf("access: %v", err)
	}
	probe := &queryOwnerScopeProbe{}
	runtime := &HTTPRuntime{
		resolver:    queryOwnerGenerationFixture{},
		coordinator: probe,
	}
	if err = call(runtime, access); err != nil {
		t.Fatalf("%s scope: %v", operation, err)
	}
	groups, domains, participants := operationAuthority(probe.request)
	if !slices.Equal(groups, []string{"analytics", "billing"}) ||
		!slices.Equal(domains, []string{"analytics", "billing"}) ||
		!slices.Equal(participants, []string{"analytics", "billing"}) {
		t.Fatalf(
			"%s authority groups=%v domains=%v participants=%v",
			operation,
			groups,
			domains,
			participants,
		)
	}
}

func operationAuthority(
	request platformtx.ScopeRequest,
) (groups, domains, participants []string) {
	operation := reflect.ValueOf(request.Operation)
	groupValues := operation.FieldByName("groups")
	for index := 0; index < groupValues.Len(); index++ {
		groups = append(
			groups,
			groupValues.Index(index).FieldByName("id").String(),
		)
	}
	domainValues := operation.FieldByName("domains")
	for index := 0; index < domainValues.Len(); index++ {
		domains = append(
			domains,
			domainValues.Index(index).FieldByName("name").String(),
		)
	}
	participantValues := operation.FieldByName("participants")
	for index := 0; index < participantValues.Len(); index++ {
		participants = append(
			participants,
			participantValues.Index(index).String(),
		)
	}
	return groups, domains, participants
}
