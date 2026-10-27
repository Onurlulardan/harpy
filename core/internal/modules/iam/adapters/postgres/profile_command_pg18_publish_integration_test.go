//go:build integration

package postgres

import (
	"context"
	"database/sql"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const callIAMProfilePublish = `
SELECT result_id, result_revision, authorization_clock, replayed
FROM iam_profile_schema_publish_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid, $2::bytea, $3::uuid, $4::text, $5::bigint,
    $6::jsonb, $7::bytea, $8::uuid, $9::uuid, $10::uuid,
    $11::bigint, $12::bigint
)`

func runIAMProfilePublish(
	ctx context.Context,
	database *sql.DB,
	tenantSchema string,
	schema profileapp.Schema,
	expectedActive uuid.UUID,
	operationID uuid.UUID,
	requestSHA256 [32]byte,
	definitionSHA256 [32]byte,
) (iamProfileCommandResult, error) {
	connection, transaction, err := beginIAMProfileCommandRuntimeTx(
		ctx, database, tenantSchema,
	)
	if err != nil {
		return iamProfileCommandResult{}, err
	}
	result, queryErr := scanIAMProfileCommand(transaction.QueryRowContext(
		ctx, callIAMProfilePublish,
		operationID, requestSHA256[:], schema.ID(), string(schema.Resource()),
		schema.Version(), string(schema.DefinitionJSON()), definitionSHA256[:],
		expectedActive, uuid.MustParse(iamUserThreeID),
		uuid.MustParse(iamRBACReadPermissionID), int64(1), int64(1),
	))
	if err = finishIAMProfileCommand(connection, transaction, queryErr); err != nil {
		return iamProfileCommandResult{}, err
	}
	return result, nil
}
