package dashboard

import (
	"context"
	"errors"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
	"testing"
)

type dashboardFixtureAuthorizer struct {
	principalErr error
	kind         string
}

func (*dashboardFixtureAuthorizer) Authorize(context.Context, platformtx.TenantScope, Permission) error {
	return nil
}
func (a *dashboardFixtureAuthorizer) AuthorizePrincipal(_ context.Context, _ platformtx.TenantScope,
	kind string, _ uuid.UUID, _ Permission,
) error {
	a.kind = kind
	return a.principalErr
}

func TestGrantACLNeverExpandsBaseIAMPermission(t *testing.T) {
	storeCalled := false
	store := &dashboardFixtureStore{grant: func(context.Context, ACLMutation) (ACL, bool, error) {
		storeCalled = true
		return ACL{}, false, nil
	}}
	authorizer := &dashboardFixtureAuthorizer{principalErr: errors.New("base permission denied")}
	application := &Application{authorizer: authorizer,
		ids:   &dashboardFixtureIDs{values: []uuid.UUID{dashboardFixtureOperation}},
		clock: dashboardFixtureClock{}}
	_, err := application.grantACLBound(context.Background(), platformtx.TenantScope{}, store,
		dashboardFixtureIdentity(), GrantACLRequest{DashboardID: dashboardFixtureID,
			PrincipalID: dashboardFixturePrincipal, PrincipalKind: "role", Level: "publish",
			IdempotencyKey: "dashboard-acl-1"}, PermissionPublish)
	if !errors.Is(err, ErrForbidden) || storeCalled || authorizer.kind != "role" {
		t.Fatalf("ACL error = %v, store = %v, kind = %s", err, storeCalled, authorizer.kind)
	}
}

func TestGrantACLAcceptsIdempotentReplay(t *testing.T) {
	store := &dashboardFixtureStore{grant: func(_ context.Context, mutation ACLMutation) (ACL, bool, error) {
		return ACL{DashboardID: mutation.DashboardID, PrincipalID: mutation.PrincipalID,
			PrincipalKind: mutation.PrincipalKind, Level: mutation.Level, Revision: 1}, true, nil
	}}
	application := &Application{authorizer: &dashboardFixtureAuthorizer{},
		entitlements: &dashboardFixtureEntitlements{decision: dashboardFixtureMutationDecision(billingapi.MeteredResourceUpdate, true)},
		ids:          &dashboardFixtureIDs{values: []uuid.UUID{dashboardFixtureOperation}},
		clock:        dashboardFixtureClock{}}
	result, err := application.grantACLBound(context.Background(), platformtx.TenantScope{}, store,
		dashboardFixtureIdentity(), GrantACLRequest{DashboardID: dashboardFixtureID,
			PrincipalID: dashboardFixturePrincipal, PrincipalKind: "user", Level: "read",
			IdempotencyKey: "dashboard-acl-2"}, PermissionRead)
	if err != nil || result.Revision != 1 {
		t.Fatalf("ACL result = %+v, err = %v", result, err)
	}
}

func TestGrantACLDoesNotWriteOwnerStateWhenBuilderIsDisabled(t *testing.T) {
	storeCalled := false
	store := &dashboardFixtureStore{grant: func(
		context.Context, ACLMutation,
	) (ACL, bool, error) {
		storeCalled = true
		return ACL{}, false, nil
	}}
	decision := dashboardFixtureMutationDecision(billingapi.MeteredResourceUpdate, false)
	decision.FeatureEnabled, decision.Allowed = false, false
	application := &Application{
		authorizer:   &dashboardFixtureAuthorizer{},
		entitlements: &dashboardFixtureEntitlements{decision: decision},
		ids: &dashboardFixtureIDs{values: []uuid.UUID{
			dashboardFixtureOperation,
		}},
		clock: dashboardFixtureClock{},
	}
	_, err := application.grantACLBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), GrantACLRequest{
			DashboardID: dashboardFixtureID, PrincipalID: dashboardFixturePrincipal,
			PrincipalKind: "role", Level: "publish",
			IdempotencyKey: "dashboard-acl-denied-1",
		}, PermissionPublish)
	if !errors.Is(err, ErrEntitlement) || storeCalled {
		t.Fatalf("grant ACL error=%v owner_called=%v", err, storeCalled)
	}
}
