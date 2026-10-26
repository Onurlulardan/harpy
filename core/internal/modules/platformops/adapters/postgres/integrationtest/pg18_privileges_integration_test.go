//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"testing"
)

func assertRecoveryPrivileges(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var tableRead, legacyClaimExecute, obsoleteClaimExecute, claimExecute bool
	var legacyRestoreExecute, restoreV2Execute, restoreV3Execute bool
	var restoreV4Execute, restoreExecute bool
	err := database.QueryRowContext(ctx, `SELECT
		has_table_privilege('harpy_backup_reader','platform.platformops_recovery_sets','SELECT'),
		has_function_privilege('harpy_backup_reader',
		 'platform.platformops_claim_recovery_set_v1(uuid,timestamptz)','EXECUTE'),
		has_function_privilege('harpy_backup_reader',
		 'platform.platformops_claim_next_recovery_set_v2(uuid,uuid)','EXECUTE'),
		has_function_privilege('harpy_backup_reader',
		 'platform.platformops_claim_next_recovery_set_v3(uuid,uuid)','EXECUTE'),
		has_function_privilege('harpy_restore',
		 'platform.platformops_claim_restore_v1(uuid,timestamptz)','EXECUTE'),
		has_function_privilege('harpy_restore',
		 'platform.platformops_claim_restore_execution_v2(uuid,uuid,uuid,text)','EXECUTE'),
		has_function_privilege('harpy_restore',
		 'platform.platformops_claim_restore_execution_v3(uuid,uuid,uuid,text)','EXECUTE'),
		has_function_privilege('harpy_restore',
		 'platform.platformops_claim_restore_execution_v4(uuid,uuid,uuid,text)','EXECUTE'),
		has_function_privilege('harpy_restore',
		 'platform.platformops_claim_restore_execution_v5(uuid,uuid,uuid,text)','EXECUTE')`).
		Scan(&tableRead, &legacyClaimExecute, &obsoleteClaimExecute, &claimExecute,
			&legacyRestoreExecute, &restoreV2Execute, &restoreV3Execute,
			&restoreV4Execute, &restoreExecute)
	if err != nil {
		t.Fatal(err)
	}
	if tableRead || legacyClaimExecute || obsoleteClaimExecute || !claimExecute ||
		legacyRestoreExecute || restoreV2Execute || restoreV3Execute ||
		restoreV4Execute || !restoreExecute {
		t.Fatalf(
			"unexpected recovery grants: table=%t legacy_claim=%t obsolete_claim=%t "+
				"claim=%t legacy_restore=%t restore_v2=%t restore_v3=%t "+
				"restore_v4=%t restore=%t",
			tableRead, legacyClaimExecute, obsoleteClaimExecute, claimExecute,
			legacyRestoreExecute, restoreV2Execute, restoreV3Execute,
			restoreV4Execute, restoreExecute,
		)
	}
}
