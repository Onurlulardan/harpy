//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

func TestContractRetirementAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	database := startCutoverIntegrationDatabase(t, ctx)
	runtimeConnection := connectContractRetirementRuntime(t, ctx, database.dsn)
	defer runtimeConnection.Close(ctx)
	authorityConnection := openCutoverMigrator(t, ctx, database.dsn)
	defer authorityConnection.Close(ctx)
	if _, err := authorityConnection.Exec(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	runtime, err := NewContractRetirementDatabase(runtimeConnection)
	if err != nil {
		t.Fatal(err)
	}
	authority, err := NewContractRetirementDatabase(authorityConnection)
	if err != nil {
		t.Fatal(err)
	}
	identity, err := contractretirement.NewIdentity("restquery", "forms")
	if err != nil {
		t.Fatal(err)
	}
	verifyContractRetirementPrivileges(t, ctx, database, runtimeConnection)
	verifyContractRetirementRuntime(t, ctx, database, runtime, authority, identity)
	verifyContractRetirementReadOnlyAdmission(
		t, ctx, database, runtimeConnection, identity,
	)
	verifyContractRetirementTransitions(t, ctx, database, runtime, authority, identity)
	verifyConcurrentContractRetirementProvision(t, ctx, database, identity)
	verifyContractRetirementConcurrency(t, ctx, database, runtime, authority, identity)
	verifyContractRetirementAdoptedMigrations(t, ctx, database)
}
