package execution

import (
	"context"
	"errors"
	"testing"

	restqueryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestExecutionRejectsTypedNilPortAndDriver(t *testing.T) {
	var nilPort testPort
	if _, err := NewRegistration("forms", nilPort); !errors.Is(err, errExecutionContract) {
		t.Fatalf("typed nil port error=%v", err)
	}
	var nilDriver *testDriver
	program := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		t.Fatal("typed nil driver reached closure")
		return nil
	}, testScopeVerifier)
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		received Program,
	) error {
		return received.Run(nilDriver)
	}))
	if err := registry.Run(program); !errors.Is(err, errExecutionContract) {
		t.Fatalf("typed nil driver error=%v", err)
	}
}

func TestExecutionRejectsWrongOwnerAtRegistryAndDriver(t *testing.T) {
	closureCalls := 0
	newProgram := func() Program {
		return newTestProgram(t, "forms", func(
			context.Context, SnapshotDriver,
		) error {
			closureCalls++
			return nil
		}, testScopeVerifier)
	}
	crmRegistry := newTestRegistry(t, "crm", testPort(func(
		transaction.TenantScope, Program,
	) error {
		t.Fatal("wrong registry port was called")
		return nil
	}))
	if err := crmRegistry.Run(newProgram()); !errors.Is(err, errExecutionContract) {
		t.Fatalf("wrong registry error=%v", err)
	}
	formsRegistry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		program Program,
	) error {
		return program.Run(&testDriver{owner: "crm"})
	}))
	if err := formsRegistry.Run(newProgram()); !errors.Is(err, errExecutionContract) {
		t.Fatalf("wrong driver error=%v", err)
	}
	if closureCalls != 0 {
		t.Fatal("wrong owner reached closure")
	}
}

func TestScopeIdentityMismatchFailsClosed(t *testing.T) {
	expected := testScopeIdentity()
	tests := map[string]scopeIdentity{
		"tenant":    withTenant(expected, uuid.MustParse("33333333-3333-4333-8333-333333333333")),
		"actor":     withActor(expected, uuid.MustParse("44444444-4444-4444-8444-444444444444")),
		"mode":      withMode(expected, transaction.Command),
		"operation": withOperation(expected, "restquery.other"),
	}
	for name, observed := range tests {
		t.Run(name, func(t *testing.T) {
			if err := verifyScopeObservation(expected, observed); !errors.Is(
				err, errExecutionContract,
			) {
				t.Fatalf("mismatch error=%v", err)
			}
		})
	}
}

func TestPublicMintRejectsUnownedContextAndScope(t *testing.T) {
	_, err := MintProgram(
		context.Background(), transaction.TenantScope{},
		restqueryapi.CatalogReference{},
		func(context.Context, SnapshotDriver) error { return nil },
	)
	if !errors.Is(err, errExecutionContract) {
		t.Fatalf("mint error=%v", err)
	}
}

func withTenant(identity scopeIdentity, value uuid.UUID) scopeIdentity {
	identity.tenantID = value
	return identity
}

func withActor(identity scopeIdentity, value uuid.UUID) scopeIdentity {
	identity.actorID = value
	return identity
}

func withMode(identity scopeIdentity, value transaction.ScopeMode) scopeIdentity {
	identity.mode = value
	return identity
}

func withOperation(identity scopeIdentity, value string) scopeIdentity {
	identity.operationID = value
	return identity
}
