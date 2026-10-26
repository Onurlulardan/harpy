//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type snapshotOwnerProbe struct {
	database postgres.BoundExecutor
}

func snapshotOwnerBinder(
	t *testing.T, owner string,
) postgres.OwnerBinder[snapshotOwnerProbe] {
	t.Helper()
	binder, err := postgres.NewOwnerBinder(
		owner,
		func(database postgres.BoundExecutor) snapshotOwnerProbe {
			return snapshotOwnerProbe{database: database}
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return binder
}

func assertMultiOwnerSnapshotAuthority(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
) {
	t.Helper()
	forms := snapshotOwnerBinder(t, "forms")
	crm := snapshotOwnerBinder(t, "crm")
	iam := snapshotOwnerBinder(t, "iam")
	for _, tenant := range []struct {
		id        string
		formsWant string
		crmWant   string
	}{
		{tenantA, "tenant-a-forms", "tenant-a-crm"},
		{tenantB, "tenant-b-forms", "tenant-b-crm"},
	} {
		request := scopeSnapshotRequestFor(
			t, tenant.id, map[string]int64{"forms": 1, "crm": 1},
		)
		err := coordinator.InTenant(
			ctx, request, func(ctx context.Context, scope TenantScope) error {
				formsProbe, err := Bind(scope, forms)
				if err != nil {
					return err
				}
				crmProbe, err := Bind(scope, crm)
				if err != nil {
					return err
				}
				if _, err := Bind(scope, iam); !errors.Is(
					err, postgres.ErrOwnerBindingUnauthorized,
				) {
					t.Fatalf("owner outside snapshot capability error = %v", err)
				}
				return assertSnapshotProbeValues(
					ctx, formsProbe, crmProbe, tenant.formsWant, tenant.crmWant,
				)
			},
		)
		if err != nil {
			t.Fatalf("multi-owner snapshot for %s: %v", tenant.id, err)
		}
	}
}

func assertSnapshotProbeValues(
	ctx context.Context,
	forms snapshotOwnerProbe,
	crm snapshotOwnerProbe,
	formsWant string,
	crmWant string,
) error {
	var formsValue, formsSnapshot string
	if err := forms.database.QueryRow(
		ctx, "SELECT value, pg_current_snapshot()::text FROM forms_probe",
	).Scan(&formsValue, &formsSnapshot); err != nil {
		return err
	}
	var crmValue, crmSnapshot string
	if err := crm.database.QueryRow(
		ctx, "SELECT value, pg_current_snapshot()::text FROM crm_probe",
	).Scan(&crmValue, &crmSnapshot); err != nil {
		return err
	}
	if formsValue != formsWant || crmValue != crmWant ||
		formsSnapshot == "" || formsSnapshot != crmSnapshot {
		return errors.New("multi-owner snapshot or tenant identity differs")
	}
	return nil
}

func assertStaleSnapshotAuthority(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
) {
	t.Helper()
	cases := []map[string]int64{
		{"crm": 2, "forms": 1},
		{"forms": 1, "missing": 1},
	}
	for index, generations := range cases {
		request := scopeSnapshotRequestFor(t, tenantA, generations)
		called := false
		err := coordinator.InTenant(
			ctx, request, func(context.Context, TenantScope) error {
				called = true
				return nil
			},
		)
		if !errors.Is(err, ErrWriterAuthority) || called {
			t.Fatalf(
				"invalid snapshot authority %d error=%v called=%t", index, err, called,
			)
		}
	}
}
