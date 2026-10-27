package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRegistryIsExactImmutableAndSupportsSameScopeOwners(t *testing.T) {
	formsCalls, crmCalls := 0, 0
	forms, _ := NewRegistration("forms", testPort(func(
		_ transaction.TenantScope,
		program Program,
	) error {
		formsCalls++
		return program.Run(&testDriver{owner: "forms"})
	}))
	crm, _ := NewRegistration("crm", testPort(func(
		_ transaction.TenantScope,
		program Program,
	) error {
		crmCalls++
		return program.Run(&testDriver{owner: "crm"})
	}))
	registrations := []Registration{forms, crm}
	registry, err := NewRegistry(registrations)
	if err != nil {
		t.Fatal(err)
	}
	registrations[0] = Registration{}
	for _, owner := range []string{"forms", "crm"} {
		program := newTestProgram(t, owner, func(
			context.Context, SnapshotDriver,
		) error {
			return nil
		}, testScopeVerifier)
		if err := registry.Run(program); err != nil {
			t.Fatalf("owner %s: %v", owner, err)
		}
	}
	if formsCalls != 1 || crmCalls != 1 {
		t.Fatalf("forms=%d crm=%d", formsCalls, crmCalls)
	}
}

func TestRegistryRejectsInvalidAndDuplicateRegistrations(t *testing.T) {
	port := testPort(func(transaction.TenantScope, Program) error { return nil })
	valid, _ := NewRegistration("forms", port)
	tests := []struct {
		name          string
		registrations []Registration
	}{
		{"empty", nil},
		{"zero", []Registration{{}}},
		{"duplicate", []Registration{valid, valid}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewRegistry(test.registrations); !errors.Is(
				err, errExecutionContract,
			) {
				t.Fatalf("error=%v", err)
			}
		})
	}
	if _, err := NewRegistration("INVALID", port); !errors.Is(
		err, errExecutionContract,
	) {
		t.Fatalf("invalid owner error=%v", err)
	}
}
