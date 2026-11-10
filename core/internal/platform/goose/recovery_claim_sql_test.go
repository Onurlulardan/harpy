package goose

import (
	"os"
	"strings"
	"testing"
)

func TestRecoveryClaimReturnsNoRowWhenIdle(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00808_platformops_claim_next_recovery_set_idle.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(body), "-- +goose Down")
	if !found || !strings.Contains(up, "WHERE false") {
		t.Fatal("idle recovery claim still returns a row")
	}
}

func TestRecoveryClaimStubStaysImmutable(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00715_platformops_claim_next_recovery_set_v3.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "WHERE false") {
		t.Fatal("applied recovery claim migration changed")
	}
}
