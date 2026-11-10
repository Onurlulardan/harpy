package recoverytarget

import (
	"context"
	"errors"
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

type failingRecoveryReader struct{ err error }

func (reader failingRecoveryReader) Get(
	context.Context, string,
) (tenantapi.RecoveryTarget, bool, error) {
	return tenantapi.RecoveryTarget{}, false, reader.err
}

func TestGetKeepsTheRecoveryTargetFailure(t *testing.T) {
	service, err := New(failingRecoveryReader{
		err: errors.New("tenancy recovery target is missing"),
	}, "backup")
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Get(context.Background())
	if !errors.Is(err, ErrUnavailable) ||
		err.Error() != "tenancy recovery target is unavailable: tenancy recovery target is missing" {
		t.Fatalf("reader failure was dropped: %v", err)
	}
}
