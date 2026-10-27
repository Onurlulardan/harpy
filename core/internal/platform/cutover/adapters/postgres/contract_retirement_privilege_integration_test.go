//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func verifyContractRetirementPrivileges(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	runtime *pgx.Conn,
) {
	t.Helper()
	for _, statement := range []string{
		`INSERT INTO platform.contract_retirements DEFAULT VALUES`,
		`UPDATE platform.contract_retirements SET revision=revision+1 WHERE false`,
		`DELETE FROM platform.contract_retirements WHERE false`,
	} {
		_, err := runtime.Exec(ctx, statement)
		var postgresError *pgconn.PgError
		if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
			t.Fatalf("runtime DML error = %v", err)
		}
	}
	var hit, admit, tableInsert, publicHit bool
	err := database.admin.QueryRow(ctx, `
SELECT has_function_privilege('harpy_runtime',
           'platform.contract_retirement_hit_v1(jsonb,text,text)', 'EXECUTE'),
       has_function_privilege('harpy_runtime',
           'platform.contract_retirement_admit_v1(jsonb,text,text)', 'EXECUTE'),
       has_table_privilege('harpy_runtime',
           'platform.contract_retirements', 'INSERT'),
       has_function_privilege('public',
           'platform.contract_retirement_hit_v1(jsonb,text,text)', 'EXECUTE')`).Scan(
		&hit, &admit, &tableInsert, &publicHit,
	)
	if err != nil || !hit || !admit || tableInsert || publicHit {
		t.Fatalf("retirement ACL = %t/%t/%t/%t err=%v",
			hit, admit, tableInsert, publicHit, err)
	}
}
