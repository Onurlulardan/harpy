//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func assertFinalizationEvidence(
	t *testing.T, ctx context.Context, admin *pgx.Conn, evidence finalizationEvidence,
) {
	t.Helper()
	var finalizations, credentials int
	var state, terminalState, terminalSHA256 string
	var catalogSHA256 *string
	var requirementCount *int
	err := admin.QueryRow(ctx, `SELECT
    (SELECT count(*) FROM cutover_authority.finalization_receipts WHERE operation_id=$1),
    (SELECT count(*) FROM platform.cutover_credential_receipts WHERE operation_id=$1),
    operation.state, receipt.tenant_requirement_catalog_sha256,
    receipt.tenant_requirement_count,receipt.terminal_state,
    receipt.terminal_evidence_sha256
FROM platform.cutover_operations AS operation
JOIN cutover_authority.finalization_receipts AS receipt
  ON receipt.operation_id=operation.operation_id AND receipt.run_id=operation.run_id
WHERE operation.operation_id=$1 AND operation.run_id=$2`,
		evidence.operationID, evidence.runID).Scan(
		&finalizations, &credentials, &state, &catalogSHA256, &requirementCount,
		&terminalState, &terminalSHA256,
	)
	validTenant := catalogSHA256 == nil && requirementCount == nil
	if evidence.terminalState == "completed" {
		_, expectedCatalog := tenantRequirementCatalogEvidence(t)
		validTenant = catalogSHA256 != nil && *catalogSHA256 == expectedCatalog &&
			requirementCount != nil && *requirementCount == 3
	}
	if err != nil || finalizations != 1 || credentials != 2 ||
		state != evidence.terminalState || terminalState != evidence.terminalState ||
		terminalSHA256 != evidence.terminalSHA256 || !validTenant {
		actualCatalog, actualCount := "<nil>", -1
		if catalogSHA256 != nil {
			actualCatalog = *catalogSHA256
		}
		if requirementCount != nil {
			actualCount = *requirementCount
		}
		t.Fatalf("finalization evidence = %d %d %s %s %d %s %s, err=%v",
			finalizations, credentials, state, actualCatalog, actualCount,
			terminalState, terminalSHA256, err)
	}
}
