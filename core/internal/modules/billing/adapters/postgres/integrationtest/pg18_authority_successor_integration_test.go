//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertBillingFinalizedAuthoritySuccessors(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	prepareBillingSuccessorFixture(t, ctx, database)
	assertBillingSuccessorPrivileges(t, ctx, database)
	assertBillingSuccessorWrongPrincipal(t, ctx, database)
	connection := openBillingFinalizerConnection(t, ctx, database)
	defer closeBillingFinalizerConnection(connection)
	assertBillingSuccessorNegativeEvidence(t, ctx, database, connection)
	assertBillingSuccessorEvidenceEquality(t, ctx, database, connection)
	globalID, err := callBillingSuccessor(
		ctx, connection, billingGlobalSuccessorOp, billingGlobalSuccessorRun,
		nil, nil, billingAcceptanceC,
	)
	if err != nil {
		t.Fatalf("finalize global Billing authority: %v", err)
	}
	replayedGlobalID, err := callBillingSuccessor(
		ctx, connection, billingGlobalSuccessorOp, billingGlobalSuccessorRun,
		nil, nil, billingAcceptanceC,
	)
	if err != nil || replayedGlobalID != globalID {
		t.Fatalf("global successor replay=%q want=%q err=%v",
			replayedGlobalID, globalID, err)
	}
	assertBillingCompetingSuccessorRejected(t, ctx, database, connection)
	tenantID := finalizeBillingTenantAuthorityRace(t, ctx, database)
	replayedTenantID, err := callBillingSuccessor(
		ctx, connection, billingTenantSuccessorOp, billingTenantSuccessorRun,
		billingFormsSuccessorOp, billingFormsSuccessorRun, billingAcceptanceC,
	)
	if err != nil || replayedTenantID != tenantID {
		t.Fatalf("tenant successor replay=%q want=%q err=%v",
			replayedTenantID, tenantID, err)
	}
	assertBillingSuccessorRows(t, ctx, database, globalID, tenantID)
}

func assertBillingCompetingSuccessorRejected(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	connection *sql.Conn,
) {
	t.Helper()
	seedBillingSuccessorOperation(
		t, ctx, database, billingCompetingGlobalOp, billingCompetingGlobalRun,
		"global", billingEnvironmentID, "billing", "billing",
	)
	_, err := callBillingSuccessor(
		ctx, connection, billingCompetingGlobalOp, billingCompetingGlobalRun,
		nil, nil, billingAcceptanceC,
	)
	requireBillingPGCode(t, err, "P5721", "competing successor operation")
}

func assertBillingSuccessorPrivileges(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var ownerExecute, finalizerExecute, wrapperExecute bool
	var finalizerGlobalInsert, ownerGlobalInsert bool
	var finalizerTenantInsert, ownerTenantInsert bool
	err := database.QueryRowContext(ctx, `
SELECT
 has_function_privilege('harpy_cutover_authority_owner',
 'platform.billing_finalize_authority_successor_v1(uuid,uuid,text)',
 'EXECUTE'),
 has_function_privilege('harpy_cutover_finalizer',
 'platform.billing_finalize_authority_successor_v1(uuid,uuid,text)',
 'EXECUTE'),
 has_function_privilege('harpy_cutover_finalizer',
 'cutover_authority.billing_test_finalize_v1(uuid,uuid,text)',
 'EXECUTE'),
 has_table_privilege('harpy_cutover_finalizer',
 'platform.billing_global_authorities','INSERT'),
 has_table_privilege('harpy_cutover_authority_owner',
 'platform.billing_global_authorities','INSERT'),
 has_table_privilege('harpy_cutover_finalizer',
 'platform.billing_tenant_authorities','INSERT'),
 has_table_privilege('harpy_cutover_authority_owner',
 'platform.billing_tenant_authorities','INSERT')`,
	).Scan(&ownerExecute, &finalizerExecute, &wrapperExecute,
		&finalizerGlobalInsert, &ownerGlobalInsert,
		&finalizerTenantInsert, &ownerTenantInsert)
	if err != nil {
		t.Fatal(err)
	}
	if !ownerExecute || finalizerExecute || !wrapperExecute ||
		finalizerGlobalInsert || ownerGlobalInsert ||
		finalizerTenantInsert || ownerTenantInsert {
		t.Fatalf("successor privilege shape = %t/%t/%t/%t/%t/%t/%t",
			ownerExecute, finalizerExecute, wrapperExecute,
			finalizerGlobalInsert, ownerGlobalInsert,
			finalizerTenantInsert, ownerTenantInsert)
	}
}

func assertBillingSuccessorWrongPrincipal(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var authorityID string
	err := database.QueryRowContext(ctx, `
SELECT platform.billing_finalize_authority_successor_v1($1,$2,$3)`,
		billingGlobalSuccessorOp, billingGlobalSuccessorRun,
		billingAcceptanceC,
	).Scan(&authorityID)
	requireBillingPGCode(t, err, "P5721", "successor wrong principal")
}

func assertBillingSuccessorNegativeEvidence(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	connection *sql.Conn,
) {
	t.Helper()
	tests := []struct {
		name, operationID, runID, digest string
		formsOperationID, formsRunID     any
	}{
		{"run", billingGlobalSuccessorOp, billingWrongSuccessorRun,
			billingAcceptanceC, nil, nil},
		{"digest", billingGlobalSuccessorOp, billingGlobalSuccessorRun,
			billingWrongSuccessorDigest, nil, nil},
		{"global companion", billingGlobalSuccessorOp, billingGlobalSuccessorRun,
			billingAcceptanceC, billingFormsSuccessorOp, billingFormsSuccessorRun},
		{"missing companion", billingTenantSuccessorOp, billingTenantSuccessorRun,
			billingAcceptanceC, nil, nil},
		{"companion run", billingTenantSuccessorOp, billingTenantSuccessorRun,
			billingAcceptanceC, billingFormsSuccessorOp, billingWrongSuccessorRun},
	}
	for _, test := range tests {
		_, err := callBillingSuccessor(ctx, connection,
			test.operationID, test.runID, test.formsOperationID,
			test.formsRunID, test.digest)
		requireBillingPGCode(t, err, "P5721", "successor wrong "+test.name)
	}
	_, err := database.ExecContext(ctx, `
INSERT INTO platform.cutover_group_members(
 scope_kind,scope_id,group_id,ordinal,domain
) VALUES('tenant',$1,'billing_forms',3,'analytics')`, billingTenantID)
	requireBillingPGCode(t, err, "23514", "cutover extra member")
}

func requireBillingPGCode(t *testing.T, err error, code, label string) {
	t.Helper()
	if billingPGCode(err) != code {
		t.Fatalf("%s code=%q want=%q err=%v", label, billingPGCode(err), code, err)
	}
}
