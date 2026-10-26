package api

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRunEnforcesManifestAndClosesScope(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	grant := validGrant(now, "environment_initialize")
	var captured Scope
	err := Run(grant, now, func(scope Scope) error {
		captured = scope
		if scope.Binding() != grant.Binding || scope.Revision() != grant.Revision {
			return errors.New("scope did not preserve the exact activation grant")
		}
		manifest := Manifest(grant.Binding.Kind)
		for _, command := range manifest[:len(manifest)-1] {
			if err := scope.Consume(command); err != nil {
				return err
			}
		}
		return nil
	})
	_, serializationErr := json.Marshal(captured)
	if err != nil || serializationErr == nil || captured.Binding() != (Binding{}) ||
		captured.Revision() != 0 || captured.OperationID() != uuid.Nil ||
		!errors.Is(captured.Consume(ActivateTenant), ErrClosed) {
		t.Fatalf("closed scope result = %v, serialization=%v, run=%v",
			captured.OperationID(), serializationErr, err)
	}
}

func TestRunRejectsOrderReplayAndExpiredLease(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	grant := validGrant(now, "saas_customer_tenant_provision")
	err := Run(grant, now, func(scope Scope) error {
		return scope.Consume(AssignTenantBaseOffering)
	})
	if !errors.Is(err, ErrOrder) {
		t.Fatalf("out-of-order error = %v", err)
	}
	grant.LeaseUntil = now
	if err := Run(grant, now, func(Scope) error { return nil }); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expired lease error = %v", err)
	}
}

func TestRunInvalidatesScopeWhenCallbackCrashes(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	var captured Scope
	crash := errors.New("crash")
	err := Run(validGrant(now, "environment_initialize"), now, func(scope Scope) error {
		captured = scope
		return crash
	})
	if !errors.Is(err, crash) || captured.OperationID() != uuid.Nil {
		t.Fatalf("callback crash = %v, scope=%v", err, captured.OperationID())
	}
}

func validGrant(now time.Time, kind string) Grant {
	environmentID, tenantID := uuid.New(), uuid.New()
	binding := Binding{OperationID: uuid.New(), EnvironmentID: environmentID,
		TenantID: tenantID, FirstUserID: uuid.New(), RootOrganizationID: uuid.New(),
		ProtectedAdminBindingID: uuid.New(), Kind: kind, DeploymentMode: "saas",
		BindingPurpose: "saas_customer", SchemaName: "t_fixture",
		RequestSHA256: strings.Repeat("1", 64), ManifestSHA256: strings.Repeat("2", 64),
		StepPlanSHA256: strings.Repeat("a", 64),
		Offering: OfferingPin{VersionID: uuid.New(), Revision: 1,
			SHA256: strings.Repeat("3", 64), EffectiveAt: now},
		Authority: AuthorityPin{RegistryBindingID: uuid.New(), ScopeKind: "tenant",
			ScopeID: tenantID, Generation: 1, StateSHA256: strings.Repeat("4", 64),
			ProposalSHA256: strings.Repeat("5", 64), AckSHA256: strings.Repeat("6", 64)},
		MigrationReceipt: MigrationReceiptPin{ID: uuid.New(), SHA256: strings.Repeat("b", 64)}}
	if kind == "environment_initialize" {
		binding.DeploymentMode, binding.BindingPurpose = "onpremise", "onpremise_primary"
		binding.InstanceOwnerBindingID = uuid.New()
		binding.Offering.Source = OfferingSourcePin{Sequence: 1,
			SHA256: strings.Repeat("7", 64), Present: true}
		binding.RecoverySigning = RecoverySigningPin{Generation: 1,
			TrustBundleSHA256: strings.Repeat("8", 64), Present: true}
		binding.Authority = AuthorityPin{RegistryBindingID: uuid.New(),
			ScopeKind: "environment", ScopeID: environmentID}
	}
	return Grant{Binding: binding, Revision: 4, LeaseUntil: now.Add(time.Minute)}
}
