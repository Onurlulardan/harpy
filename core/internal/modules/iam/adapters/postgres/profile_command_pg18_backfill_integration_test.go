//go:build integration

package postgres

import (
	"context"
	"database/sql"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const callIAMProfileBackfill = `
SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_profile_document_backfill_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::text, $4::uuid, $5::uuid, $6::uuid,
    $7::bytea, $8::bytea, $9::bigint, $10::uuid, $11::uuid,
    $12::bigint, $13::bigint
)`

func runIAMProfileBackfill(
	ctx context.Context,
	database *sql.DB,
	tenantSchema string,
	resourceID uuid.UUID,
	expectedSource uuid.UUID,
	expectedRevision int64,
	document profileapp.Document,
	operationID uuid.UUID,
	requestSHA256 [32]byte,
) (iamProfileCommandResult, error) {
	connection, transaction, err := beginIAMProfileCommandRuntimeTx(
		ctx, database, tenantSchema,
	)
	if err != nil {
		return iamProfileCommandResult{}, err
	}
	documentSHA256 := document.SHA256()
	result, queryErr := scanIAMProfileCommand(transaction.QueryRowContext(
		ctx, callIAMProfileBackfill,
		operationID, requestSHA256[:], string(document.Resource()), resourceID,
		expectedSource, document.SchemaID(), document.CanonicalJSON(),
		documentSHA256[:], expectedRevision, uuid.MustParse(iamUserThreeID),
		uuid.MustParse(iamRBACReadPermissionID), int64(1), int64(1),
	))
	if err = finishIAMProfileCommand(connection, transaction, queryErr); err != nil {
		return iamProfileCommandResult{}, err
	}
	return result, nil
}
