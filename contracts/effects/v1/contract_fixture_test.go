package effectsv1

import (
	"strings"
	"testing"
)

const (
	testTenantID  = "11111111-1111-4111-8111-111111111111"
	testEffectID  = "22222222-2222-4222-8222-222222222222"
	testReceiptID = "33333333-3333-4333-8333-333333333333"
	testGroupID   = "notifications"
	testCauseID   = "55555555-5555-4555-8555-555555555555"
)

func requestFixture(t *testing.T) Request {
	t.Helper()
	tenantID := testTenantID
	value, err := NewRequest(
		testEffectID,
		"notifications.email.send",
		1,
		&tenantID,
		Authority{
			ScopeKind: ScopeTenant, ScopeID: testTenantID, Domain: "notifications",
			GroupID: testGroupID, GroupGeneration: 3,
			CutoverGeneration: 7, WriterOrigin: WriterV2,
		},
		"email-provider",
		"send",
		"effect:notification:42",
		"request:notification.42",
		testCauseID,
		"notification-delivery/42",
		strings.Repeat("a", 64),
		"2026-07-15T10:00:00Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func successReceiptFixture(t *testing.T) Receipt {
	t.Helper()
	providerReference := "provider-message/99"
	value, err := NewSucceededReceipt(
		requestFixture(t), testReceiptID, "DELIVERED", &providerReference,
		"notification-receipt/42", strings.Repeat("b", 64),
		"2026-07-15T10:00:01Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
