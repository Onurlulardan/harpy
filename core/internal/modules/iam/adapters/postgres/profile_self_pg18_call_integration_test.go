//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const callIAMSelfProfileRead = `
SELECT result_id,email,status,revision,created_at,updated_at,roles,organizations
FROM iam_self_profile_read_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3)`

const callIAMSelfProfileUpdate = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_self_profile_update_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`

type iamSelfProfileRead struct {
	id            uuid.UUID
	email         string
	status        string
	revision      int64
	createdAt     time.Time
	updatedAt     time.Time
	roles         []byte
	organizations []byte
}

func runIAMSelfProfileRead(
	ctx context.Context,
	database *sql.DB,
	schema string,
	actorID uuid.UUID,
	tokenVersion int64,
	authorizationVersion int64,
) (iamSelfProfileRead, error) {
	connection, transaction, err := beginIAMProfileCommandRuntimeTx(
		ctx, database, schema,
	)
	if err != nil {
		return iamSelfProfileRead{}, err
	}
	var result iamSelfProfileRead
	queryErr := transaction.QueryRowContext(
		ctx, callIAMSelfProfileRead, actorID, tokenVersion, authorizationVersion,
	).Scan(
		&result.id, &result.email, &result.status, &result.revision,
		&result.createdAt, &result.updatedAt, &result.roles, &result.organizations,
	)
	if err = finishIAMProfileCommand(connection, transaction, queryErr); err != nil {
		return iamSelfProfileRead{}, err
	}
	return result, nil
}

func runIAMSelfProfileUpdate(
	ctx context.Context,
	database *sql.DB,
	schema string,
	operationID uuid.UUID,
	requestSHA [32]byte,
	actorID uuid.UUID,
	sourceSchemaID uuid.UUID,
	targetSchemaID uuid.UUID,
	document []byte,
	documentSHA [32]byte,
	expected iamIdentityActor,
) (iamProfileCommandResult, error) {
	connection, transaction, err := beginIAMProfileCommandRuntimeTx(
		ctx, database, schema,
	)
	if err != nil {
		return iamProfileCommandResult{}, err
	}
	result, queryErr := scanIAMProfileCommand(transaction.QueryRowContext(
		ctx, callIAMSelfProfileUpdate,
		operationID, requestSHA[:], actorID, sourceSchemaID, targetSchemaID,
		document, documentSHA[:], expected.revision,
		expected.tokenVersion, expected.authorizationVersion,
	))
	if err = finishIAMProfileCommand(connection, transaction, queryErr); err != nil {
		return iamProfileCommandResult{}, err
	}
	return result, nil
}
