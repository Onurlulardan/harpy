package recoverytarget

import (
	"context"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type readerFixture struct {
	purpose string
	target  tenantapi.RecoveryTarget
}

func (fixture *readerFixture) Get(
	_ context.Context, purpose string,
) (tenantapi.RecoveryTarget, bool, error) {
	fixture.purpose = purpose
	return fixture.target, true, nil
}

func TestGetUsesConfiguredRecoveryPurpose(t *testing.T) {
	target, err := tenantapi.NewRecoveryTarget(
		"10000000-0000-4000-8000-000000000001",
		"20000000-0000-4000-8000-000000000002",
		"primary", "tenant_primary", [32]byte{1},
	)
	if err != nil {
		t.Fatalf("target: %v", err)
	}
	fixture := &readerFixture{target: target}
	service, err := New(fixture, "restore")
	if err != nil {
		t.Fatalf("service: %v", err)
	}
	if _, err = service.Get(context.Background()); err != nil {
		t.Fatalf("get: %v", err)
	}
	if fixture.purpose != "restore" {
		t.Fatalf("unexpected purpose: %s", fixture.purpose)
	}
}
