//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

type billingCapabilityFixture struct {
	operationID, environmentID, tenantID uuid.UUID
	authorityID, assignmentID            uuid.UUID
	offering                             billingInstallationOffering
	manifestSHA256                       string
}

type billingCapabilityMint struct {
	operationKind, deploymentMode, bindingPurpose string
	manifestSHA256                                string
	offeringVersionID                             uuid.UUID
	sourceSequence, sourceSHA256                  any
}

type billingCapabilityNegative struct {
	name, mode, purpose, want string
	authorityID, assignmentID uuid.UUID
	mint                      *billingCapabilityMint
}

func assertBillingInstallationCapabilityNegatives(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	fixture := loadBillingCapabilityFixture(t, ctx, database)
	valid := billingCapabilityMint{
		operationKind: "environment_initialize", deploymentMode: "saas",
		bindingPurpose: "saas_control", manifestSHA256: fixture.manifestSHA256,
		offeringVersionID: fixture.offering.versionID,
	}
	wrongKind := valid
	wrongKind.operationKind = "saas_customer_tenant_provision"
	wrongManifest := valid
	wrongManifest.manifestSHA256 = billingAcceptanceB
	wrongOffering := valid
	wrongOffering.offeringVersionID = billingV7(t)
	wrongSource := valid
	wrongSource.sourceSequence = int64(1)
	wrongSource.sourceSHA256 = billingAcceptanceB
	changedProfile := valid
	changedProfile.operationKind = "saas_customer_tenant_provision"
	changedProfile.bindingPurpose = "saas_customer"
	tests := []billingCapabilityNegative{
		{"random_authority", "saas", "saas_control", "42501",
			billingV7(t), fixture.assignmentID, &valid},
		{"random_assignment", "saas", "saas_control", "42501",
			fixture.authorityID, billingV7(t), &valid},
		{"onpremise_control", "onpremise", "saas_control", "22023",
			fixture.authorityID, fixture.assignmentID, nil},
		{"saas_onpremise", "saas", "onpremise_primary", "22023",
			fixture.authorityID, fixture.assignmentID, nil},
		{"wrong_operation_kind", "saas", "saas_control", "42501",
			fixture.authorityID, fixture.assignmentID, &wrongKind},
		{"wrong_manifest", "saas", "saas_control", "42501",
			fixture.authorityID, fixture.assignmentID, &wrongManifest},
		{"wrong_offering", "saas", "saas_control", "42501",
			fixture.authorityID, fixture.assignmentID, &wrongOffering},
		{"wrong_source", "saas", "saas_control", "42501",
			fixture.authorityID, fixture.assignmentID, &wrongSource},
		{"changed_profile", "saas", "saas_customer", "23505",
			fixture.authorityID, fixture.assignmentID, &changedProfile},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := callBillingCapabilityAttempt(
				t, ctx, database, fixture, test.authorityID,
				test.assignmentID, test.mode, test.purpose, test.mint,
			)
			requireBillingCapabilityCode(t, err, test.want)
		})
	}
	assertBillingInstallationContextImmutable(t, ctx, database, fixture.operationID)
}

func loadBillingCapabilityFixture(
	t *testing.T, ctx context.Context, database *sql.DB,
) billingCapabilityFixture {
	t.Helper()
	var fixture billingCapabilityFixture
	err := database.QueryRowContext(ctx, `
SELECT fact.operation_id,fact.environment_id,fact.tenant_id,
 fact.authority_id,fact.assignment_id,fact.offering_version_id,
 fact.offering_revision,fact.offering_sha256,fact.offering_effective_at,
 fact.manifest_sha256
FROM platform.billing_installation_base_assignments AS fact
JOIN platform.billing_installation_assignment_contexts AS context
  ON context.operation_id=fact.operation_id
WHERE context.deployment_mode='saas' AND context.operation_kind='environment_initialize'
  AND context.binding_purpose='saas_control'
ORDER BY fact.assigned_at LIMIT 1`).Scan(
		&fixture.operationID, &fixture.environmentID, &fixture.tenantID,
		&fixture.authorityID, &fixture.assignmentID, &fixture.offering.versionID,
		&fixture.offering.revision, &fixture.offering.sha256,
		&fixture.offering.effective, &fixture.manifestSHA256,
	)
	if err != nil {
		t.Fatalf("load Billing capability fixture: %v", err)
	}
	return fixture
}

func callBillingCapabilityAttempt(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture billingCapabilityFixture, authorityID, assignmentID uuid.UUID,
	mode, purpose string, mint *billingCapabilityMint,
) error {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mint != nil {
		mintBillingCapabilityAttempt(t, ctx, transaction, fixture, *mint)
	}
	if _, err = transaction.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_owner;
SELECT pg_catalog.set_config(
 'application_name','harpy-v2:core-operator-initialize',true
)`); err != nil {
		t.Fatal(err)
	}
	var replayed bool
	err = transaction.QueryRowContext(ctx, `
SELECT platform.billing_installation_assign_v3(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULL,NULL,$11,$12
)`, fixture.operationID, fixture.environmentID, fixture.tenantID,
		authorityID, assignmentID, mode, fixture.offering.versionID,
		fixture.offering.revision, fixture.offering.sha256,
		fixture.offering.effective, fixture.manifestSHA256, purpose,
	).Scan(&replayed)
	_ = transaction.Rollback()
	_, _ = connection.ExecContext(context.Background(),
		"RESET ROLE; RESET SESSION AUTHORIZATION")
	return err
}

func mintBillingCapabilityAttempt(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
	fixture billingCapabilityFixture, mint billingCapabilityMint,
) {
	t.Helper()
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.test_tenancy_activation_commands(
 backend_pid,transaction_id,operation_id,command_name,environment_id,tenant_id,
 operation_kind,deployment_mode,binding_purpose,manifest_sha256,
 offering_version_id,offering_revision,offering_sha256,offering_effective_at,
 source_sequence,source_sha256
) VALUES(
 pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),$1,
 'billing.assign_tenant_base_offering',$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13
)`, fixture.operationID, fixture.environmentID, fixture.tenantID,
		mint.operationKind, mint.deploymentMode, mint.bindingPurpose,
		mint.manifestSHA256, mint.offeringVersionID, fixture.offering.revision,
		fixture.offering.sha256, fixture.offering.effective,
		mint.sourceSequence, mint.sourceSHA256)
	if err != nil {
		t.Fatalf("mint Billing capability attempt: %v", err)
	}
}

func requireBillingCapabilityCode(t *testing.T, err error, want string) {
	t.Helper()
	if code := billingPGCode(err); code != want {
		t.Fatalf("Billing capability code=%q want=%q err=%v", code, want, err)
	}
}

func assertBillingInstallationContextImmutable(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID uuid.UUID,
) {
	t.Helper()
	for _, statement := range []string{
		`UPDATE platform.billing_installation_assignment_contexts
SET recorded_at=recorded_at WHERE operation_id=$1`,
		`DELETE FROM platform.billing_installation_assignment_contexts
WHERE operation_id=$1`,
	} {
		_, err := database.ExecContext(ctx, statement, operationID)
		requireBillingCapabilityCode(t, err, "55000")
	}
}
