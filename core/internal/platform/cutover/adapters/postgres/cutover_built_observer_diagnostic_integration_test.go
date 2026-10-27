//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5/pgconn"
)

const builtFinalizerObserverSQL = `SELECT pg_catalog.clock_timestamp(),
    observation.finalization_state_sha256,
    observation.finalization_procedure_oid,
    observation.finalization_procedure_signature,
    observation.database_name, observation.database_origin,
    observation.system_identifier, observation.database_oid,
    observation.operation_state, observation.completed_eligible,
    observation.abort_eligible,
    observation.completed_tenant_catalog_sha256,
    observation.completed_tenant_requirement_count,
    observation.finalizer_valid_until
FROM cutover_authority.observe_finalizer_authority_v1($1::uuid,$2::uuid) AS observation`

func diagnoseBuiltFinalizerObserver(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts, seal, cleanup builtCutoverArtifact,
	expectedExpiry time.Time,
) string {
	t.Helper()
	connection := connectCutoverRole(t, ctx, database.dsn, "harpy_cluster_admin")
	defer connection.Close(context.Background())
	if err := recordCredentialReceipts(ctx, connection, adoption, seal, cleanup); err != nil {
		return fmt.Sprintf("record_error=%s", postgresDiagnostic(err))
	}
	var observed builtFinalizerObservationDiagnostic
	err := connection.QueryRow(
		ctx, builtFinalizerObserverSQL,
		adoption.identity.operationID, adoption.identity.runID,
	).Scan(
		&observed.databaseTime, &observed.stateSHA256, &observed.procedureOID,
		&observed.procedureSignature, &observed.target.DatabaseName,
		&observed.target.DatabaseOrigin, &observed.target.SystemIdentifier,
		&observed.target.DatabaseOID, &observed.operationState,
		&observed.completedEligible, &observed.abortEligible,
		&observed.tenantCatalogSHA256, &observed.tenantRequirementCount,
		&observed.finalizerValidUntil,
	)
	if err != nil {
		return "query_or_scan_error=" + postgresDiagnostic(err)
	}
	return fmt.Sprintf(
		"target_match=%t state_digest_present=%t procedure_oid=%d signature=%q "+
			"state=%q completed=%t abort=%t tenant_digest=%v tenant_count=%v "+
			"expiry=%s expected_expiry=%s expiry_match=%t remaining=%s",
		observed.target == adoption.lease.Target, observed.stateSHA256 != "",
		observed.procedureOID, observed.procedureSignature, observed.operationState,
		observed.completedEligible, observed.abortEligible,
		observed.tenantCatalogSHA256, observed.tenantRequirementCount,
		observed.finalizerValidUntil.Format(time.RFC3339Nano),
		expectedExpiry.Format(time.RFC3339Nano),
		observed.finalizerValidUntil.Equal(expectedExpiry),
		observed.finalizerValidUntil.Sub(observed.databaseTime),
	)
}

type builtFinalizerObservationDiagnostic struct {
	databaseTime           time.Time
	stateSHA256            string
	procedureOID           uint32
	procedureSignature     string
	target                 cutoverv1.DatabaseIdentity
	operationState         string
	completedEligible      bool
	abortEligible          bool
	tenantCatalogSHA256    *string
	tenantRequirementCount *int
	finalizerValidUntil    time.Time
}

func postgresDiagnostic(err error) string {
	var value *pgconn.PgError
	if !errors.As(err, &value) {
		return err.Error()
	}
	return fmt.Sprintf(
		"sqlstate=%s message=%q detail=%q hint=%q where=%q internal=%q",
		value.Code, value.Message, value.Detail, value.Hint,
		value.Where, value.InternalQuery,
	)
}
