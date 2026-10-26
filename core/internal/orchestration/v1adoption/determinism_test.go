package v1adoption

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"
)

func TestEvaluateIsDeterministicAcrossInventoryOrder(t *testing.T) {
	t.Parallel()
	control := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	customer := uuid.MustParse("33333333-3333-7333-8333-333333333333")
	inventory := validSaaSInventory(control, customer)
	inventory.Protected = append(inventory.Protected, ProtectedAccountEvidence{
		TenantID: customer,
		UserID:   uuid.MustParse("44444444-4444-7444-8444-444444444444"),
		Active:   true,
	})
	reordered := inventory
	reordered.Tenants = append([]TenantEvidence(nil), inventory.Tenants...)
	reordered.Protected = append(
		[]ProtectedAccountEvidence(nil), inventory.Protected...,
	)
	reordered.Billing = append([]BillingEvidence(nil), inventory.Billing...)
	slices.Reverse(reordered.Tenants)
	slices.Reverse(reordered.Protected)
	slices.Reverse(reordered.Billing)

	first, err := Evaluate(inventory)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Evaluate(reordered)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("first=%#v second=%#v", first, second)
	}
}
