package main

import (
	"context"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

const finalizationProcedureSignature = "cutover_authority.finalize_operation_v1(" +
	"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)"

const observeFinalizerAuthoritySQL = `SELECT pg_catalog.clock_timestamp(),
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

type finalizerAuthorityObservation struct {
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

func observeFinalizerAuthority(
	ctx context.Context, transaction pgx.Tx, operationID, runID string,
) (finalizerAuthorityObservation, error) {
	var value finalizerAuthorityObservation
	err := transaction.QueryRow(ctx, observeFinalizerAuthoritySQL, operationID, runID).Scan(
		&value.databaseTime, &value.stateSHA256, &value.procedureOID,
		&value.procedureSignature, &value.target.DatabaseName,
		&value.target.DatabaseOrigin, &value.target.SystemIdentifier,
		&value.target.DatabaseOID, &value.operationState, &value.completedEligible,
		&value.abortEligible, &value.tenantCatalogSHA256,
		&value.tenantRequirementCount, &value.finalizerValidUntil,
	)
	if err != nil {
		return finalizerAuthorityObservation{},
			fmt.Errorf("protected finalizer authority observation failed")
	}
	return value, nil
}

func (value finalizerAuthorityObservation) validate(
	target cutoverv1.DatabaseIdentity, terminalState string,
	connectedFinalizerValidUntil time.Time,
) error {
	remaining := value.finalizerValidUntil.Sub(value.databaseTime)
	if value.target != target || value.stateSHA256 == "" || value.procedureOID == 0 ||
		value.procedureSignature != finalizationProcedureSignature ||
		!value.finalizerValidUntil.Equal(connectedFinalizerValidUntil) ||
		remaining < 15*time.Second || remaining > 120*time.Second ||
		(value.tenantCatalogSHA256 == nil) != (value.tenantRequirementCount == nil) {
		return fmt.Errorf("observed finalizer authority differs")
	}
	if terminalState == "completed" {
		if !value.completedEligible || value.tenantCatalogSHA256 == nil ||
			value.tenantRequirementCount == nil {
			return fmt.Errorf("completed finalization authority is not eligible")
		}
		return nil
	}
	if !value.abortEligible {
		return fmt.Errorf("abort finalization authority is not eligible")
	}
	return nil
}
