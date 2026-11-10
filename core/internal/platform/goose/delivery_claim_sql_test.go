package goose

import (
	"os"
	"strings"
	"testing"
)

func TestDeliveryClaimReturnsNoRowWhenIdle(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00804_platformauthority_delivery_claim_idle.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(body), "-- +goose Down")
	if !found || !strings.Contains(up, "WHERE false") {
		t.Fatal("idle delivery claim still returns a row")
	}
}

func TestDeliveryClaimStubStaysImmutable(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00699_platformauthority_delivery_claim_v1.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "WHERE false") {
		t.Fatal("applied delivery claim migration changed")
	}
}
