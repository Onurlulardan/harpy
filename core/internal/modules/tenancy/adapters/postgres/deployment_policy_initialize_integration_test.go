//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"
)

const policyInitializeOperationID = "018f0000-0000-7000-8000-000000000810"

type policyInitializeResult struct {
	operationID, environmentID, mode string
	maximum                          sql.NullInt64
	replayed                         bool
}

func TestDeploymentPolicyInitializationAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			resetTenantMigrationCatalog(t, ctx, database)
			if origin == "v1_adopted" {
				prepareV1TenantRegistry(t, ctx, database)
			}
			runPlatformMigrations(t, ctx, database, origin)
			first := initializePolicyAsOperator(t, ctx, database,
				policyInitializeOperationID, "onpremise")
			if first.replayed || first.operationID != policyInitializeOperationID ||
				first.environmentID != testEnvironmentID || first.mode != "onpremise" ||
				!first.maximum.Valid || first.maximum.Int64 != 0 {
				t.Fatalf("first initialization = %#v", first)
			}
			replay := initializePolicyAsOperator(t, ctx, database,
				policyInitializeOperationID, "onpremise")
			if !replay.replayed {
				t.Fatalf("exact replay = %#v", replay)
			}
			assertPolicyInitializeDenied(t, ctx, database)
		})
	}
}

func TestDeploymentPolicyInitializationConcurrentReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startTenantMigrationDatabase(t, ctx)
	resetTenantMigrationCatalog(t, ctx, database)
	runPlatformMigrations(t, ctx, database, "fresh")
	results := make(chan policyInitializeResult, 2)
	errors := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			result, err := initializePolicyOperator(ctx, database,
				policyInitializeOperationID, "saas")
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}
	group.Wait()
	close(results)
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
	created, replayed := 0, 0
	for result := range results {
		if result.maximum.Valid {
			t.Fatalf("saas maximum = %#v", result.maximum)
		}
		if result.replayed {
			replayed++
		} else {
			created++
		}
	}
	if created != 1 || replayed != 1 {
		t.Fatalf("created=%d replayed=%d", created, replayed)
	}
}

func initializePolicyAsOperator(t *testing.T, ctx context.Context, database *sql.DB,
	operationID, mode string,
) policyInitializeResult {
	t.Helper()
	result, err := initializePolicyOperator(ctx, database, operationID, mode)
	if err != nil {
		t.Fatalf("initialize deployment policy: %v", err)
	}
	return result
}

func initializePolicyOperator(ctx context.Context, database *sql.DB,
	operationID, mode string,
) (policyInitializeResult, error) {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return policyInitializeResult{}, err
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime;
SET LOCAL ROLE harpy_runtime;
SET LOCAL application_name = 'harpy-v2:core-operator-initialize'`); err != nil {
		return policyInitializeResult{}, err
	}
	var result policyInitializeResult
	var createdAt, databaseNow time.Time
	err = transaction.QueryRowContext(ctx, `SELECT operation_id,environment_id,
deployment_mode,maximum_customer_tenants,created_at,database_now,replayed
FROM platform.tenancy_initialize_deployment_policy_v1($1,$2,'environment_initialize',$3)`,
		operationID, testEnvironmentID, mode).Scan(&result.operationID,
		&result.environmentID, &result.mode, &result.maximum, &createdAt,
		&databaseNow, &result.replayed)
	if err != nil || databaseNow.Before(createdAt) {
		return policyInitializeResult{}, err
	}
	if err = transaction.Commit(); err != nil {
		return policyInitializeResult{}, err
	}
	return result, nil
}

func assertPolicyInitializeDenied(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	_, err := database.ExecContext(ctx, `SELECT * FROM
platform.tenancy_initialize_deployment_policy_v1($1,$2,'environment_initialize','onpremise')`,
		policyInitializeOperationID, testEnvironmentID)
	if err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("owner invocation error = %v", err)
	}
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	_, err = transaction.ExecContext(ctx, `INSERT INTO
platform.tenancy_deployment_policy_initializations
(operation_id,environment_id,operation_kind,deployment_mode,maximum_customer_tenants)
VALUES ($1,$2,'environment_initialize','onpremise',0)`,
		"018f0000-0000-7000-8000-000000000811", testEnvironmentID)
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("runtime direct initialization write error = %v", err)
	}
}
