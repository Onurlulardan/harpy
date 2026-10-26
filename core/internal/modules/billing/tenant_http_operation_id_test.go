package billing

import (
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestOperationIdentityIsStableScopedUUIDv7(t *testing.T) {
	fixture := newApplicationFixture(t)
	first, err := operationIdentity(fixture.access, "billing.discount.redeem", "stable-key")
	if err != nil || first.Version() != 7 {
		t.Fatalf("operation identity=%s err=%v", first, err)
	}
	if first != mustOperationIdentity(t, fixture.access, "billing.discount.redeem", "stable-key") {
		t.Fatal("same idempotency identity changed UUIDv7 mapping")
	}
	if first == mustOperationIdentity(t, fixture.access, "billing.discount.redeem", "other-key") {
		t.Fatal("different idempotency key reused operation identity")
	}
	other, _ := billingapi.NewTenantAccess(
		fixture.access.PrincipalID(), testV7(t), 1, billingapi.PermissionBillingManage,
	)
	if first == mustOperationIdentity(t, other, "billing.discount.redeem", "stable-key") {
		t.Fatal("different tenant reused operation identity")
	}
}

func mustOperationIdentity(
	t *testing.T, access billingapi.Access, operation, key string,
) [16]byte {
	t.Helper()
	value, err := operationIdentity(access, operation, key)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
