package billing

import (
	"context"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type testResolver struct{ calls int }

func (resolver *testResolver) ResolveTenantBillingGenerations(
	_ context.Context, _ uuid.UUID,
) (int64, int64, int64, error) {
	resolver.calls++
	return 7, 8, 9, nil
}

type testCoordinator struct{ modes []platformtx.ScopeMode }

func (coordinator *testCoordinator) InTenant(
	ctx context.Context,
	request platformtx.ScopeRequest,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	coordinator.modes = append(coordinator.modes, request.Mode)
	return action(ctx, platformtx.TenantScope{})
}

type testBindings struct {
	current     billingapi.TenantSubscriptionSnapshot
	usage       billingapi.TenantUsageSnapshot
	eligibility billingapi.TenantDiscountSnapshot
	change      billingapi.TenantSubscriptionChange
	cancel      billingapi.TenantSubscriptionCancellation
	resume      billingapi.TenantSubscriptionResume
	redeem      billingapi.TenantDiscountRedemption
	lookup      billingapi.TenantDiscountLookup
	resolveCode string
	changeView  *billingapi.TenantSubscriptionSnapshot
}

func (bindings *testBindings) SubscriptionQueries(
	platformtx.TenantScope,
) (billingapi.TenantSubscriptionQueryPort, error) {
	return bindings, nil
}
func (bindings *testBindings) UsageQueries(
	platformtx.TenantScope,
) (billingapi.TenantUsageQueryPort, error) {
	return bindings, nil
}
func (bindings *testBindings) DiscountQueries(
	platformtx.TenantScope,
) (billingapi.TenantDiscountQueryPort, error) {
	return bindings, nil
}
func (bindings *testBindings) SubscriptionCommands(
	platformtx.TenantScope,
) (billingapi.TenantSubscriptionCommandPort, error) {
	return bindings, nil
}
func (bindings *testBindings) DiscountCommands(
	platformtx.TenantScope,
) (billingapi.TenantDiscountCommandPort, error) {
	return bindings, nil
}

func (bindings *testBindings) GetCurrent(
	context.Context, platformtx.TenantScope,
) (billingapi.TenantSubscriptionSnapshot, error) {
	return bindings.current, nil
}
func (bindings *testBindings) Get(
	context.Context, platformtx.TenantScope,
) (billingapi.TenantUsageSnapshot, error) {
	return bindings.usage, nil
}
func (bindings *testBindings) Validate(
	_ context.Context, _ platformtx.TenantScope, request billingapi.TenantDiscountLookup,
) (billingapi.TenantDiscountSnapshot, error) {
	bindings.lookup = request
	return bindings.eligibility, nil
}
func (bindings *testBindings) ResolveCurrent(
	_ context.Context, _ platformtx.TenantScope, code string,
) (billingapi.TenantDiscountSnapshot, error) {
	bindings.resolveCode = code
	return bindings.eligibility, nil
}

func (bindings *testBindings) ChangeAtPeriodEnd(
	_ context.Context, request billingapi.TenantSubscriptionChange,
) error {
	bindings.change = request
	if bindings.changeView != nil {
		bindings.current = *bindings.changeView
	} else {
		bindings.current = revisedSubscription(bindings.current, false)
	}
	return nil
}

func (bindings *testBindings) CancelAtPeriodEnd(
	_ context.Context, request billingapi.TenantSubscriptionCancellation,
) error {
	bindings.cancel = request
	bindings.current = revisedSubscription(bindings.current, true)
	return nil
}

func (bindings *testBindings) ResumeCancellation(
	_ context.Context, request billingapi.TenantSubscriptionResume,
) error {
	bindings.resume = request
	bindings.current = revisedSubscription(bindings.current, false)
	return nil
}

func (bindings *testBindings) Redeem(
	_ context.Context, request billingapi.TenantDiscountRedemption,
) error {
	bindings.redeem = request
	return nil
}

type applicationFixture struct {
	application *tenantOperations
	access      billingapi.Access
	bindings    *testBindings
	coordinator *testCoordinator
	resolver    *testResolver
}

func newApplicationFixture(t *testing.T) applicationFixture {
	t.Helper()
	tenantID, principalID := uuid.New(), uuid.New()
	access, err := billingapi.NewTenantAccessEvidence(
		principalID, tenantID, 2, 3, 4, billingapi.PermissionBillingManage,
	)
	if err != nil {
		t.Fatal(err)
	}
	bindings := &testBindings{current: currentSubscription(t, tenantID)}
	bindings.usage = usageFixture(tenantID)
	bindings.eligibility = eligibilitySnapshot(t, tenantID, bindings.current)
	coordinator, resolver := &testCoordinator{}, &testResolver{}
	application, err := newTenantOperations(resolver, coordinator, bindings)
	if err != nil {
		t.Fatal(err)
	}
	return applicationFixture{application, access, bindings, coordinator, resolver}
}

func testV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func accessFor(
	t *testing.T,
	base billingapi.Access,
	permission billingapi.Permission,
) billingapi.Access {
	t.Helper()
	value, err := billingapi.NewTenantAccessEvidence(
		base.PrincipalID(), base.TenantID(), base.TokenVersion(),
		base.AuthorizationVersion(), base.AuthorizationClock(), permission,
	)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
