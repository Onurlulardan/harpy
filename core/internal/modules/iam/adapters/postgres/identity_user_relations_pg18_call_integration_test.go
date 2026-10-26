//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

const iamUserRelationsReplaceSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_user_relations_replace_v1(
    $1::name,$2::uuid,$3::text,$4::bytea,$5::uuid,$6::bigint,
    $7::boolean,$8::uuid,$9::bytea,$10::bytea,
    $11::boolean,$12::uuid[],$13::uuid[],
    $14::boolean,$15::uuid[],$16::uuid[],
    $17::uuid,$18::uuid,$19::bigint,$20::bigint)`

type iamUserRelationsInput struct {
	scope, operationID, commandKey, targetID string
	requestSHA                               []byte
	expectedRevision                         int64
	patchProfile                             bool
	profileSchemaID                          string
	profileDocument, profileSHA              []byte
	replaceOrganizations                     bool
	organizationIDs, membershipIDs           []uuid.UUID
	replaceRoles                             bool
	roleIDs, assignmentIDs                   []uuid.UUID
	actorID, permissionID                    string
	tokenVersion, authorizationVersion       int64
}

type iamUserRelationsResult struct {
	id       string
	revision int64
	clock    int64
	replayed bool
}

type iamUserRelationsState struct {
	revision, authorizationVersion int64
	organizations, roles           string
}

func newIAMUserRelationsInput(
	actor iamIdentityActor, targetID, operationID, commandKey string,
	expectedRevision int64, marker byte,
) iamUserRelationsInput {
	return iamUserRelationsInput{
		scope: "t_iam_a", operationID: operationID, commandKey: commandKey,
		targetID: targetID, requestSHA: bytes.Repeat([]byte{marker}, 32),
		expectedRevision: expectedRevision, actorID: iamIdentityActorID,
		permissionID: iamIdentityPermissionID, tokenVersion: actor.tokenVersion,
		authorizationVersion: actor.authorizationVersion,
	}
}

func iamUserRelationUUIDs(values ...string) []uuid.UUID {
	result := make([]uuid.UUID, len(values))
	for index, value := range values {
		result[index] = uuid.MustParse(value)
	}
	return result
}

func queryIAMUserRelations(
	ctx context.Context, transaction *sql.Tx, input iamUserRelationsInput,
) (iamUserRelationsResult, error) {
	var result iamUserRelationsResult
	err := transaction.QueryRowContext(ctx, iamUserRelationsReplaceSQL,
		input.scope, uuid.MustParse(input.operationID), input.commandKey,
		input.requestSHA, uuid.MustParse(input.targetID), input.expectedRevision,
		input.patchProfile, nullableIAMRelationUUID(input.profileSchemaID),
		input.profileDocument, input.profileSHA,
		input.replaceOrganizations, input.organizationIDs, input.membershipIDs,
		input.replaceRoles, input.roleIDs, input.assignmentIDs,
		uuid.MustParse(input.actorID), uuid.MustParse(input.permissionID),
		input.tokenVersion, input.authorizationVersion,
	).Scan(&result.id, &result.revision, &result.clock, &result.replayed)
	return result, err
}

func nullableIAMRelationUUID(value string) any {
	if value == "" {
		return nil
	}
	return uuid.MustParse(value)
}

func executeIAMUserRelations(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
	input iamUserRelationsInput,
) iamUserRelationsResult {
	t.Helper()
	var result iamUserRelationsResult
	withIAMIdentityRuntimeTx(t, ctx, database, schema, func(transaction *sql.Tx) {
		var err error
		result, err = queryIAMUserRelations(ctx, transaction, input)
		if err != nil {
			t.Fatalf("replace IAM user relations: %v", err)
		}
	})
	return result
}

func rejectIAMUserRelations(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
	input iamUserRelationsInput,
) error {
	t.Helper()
	var queryErr error
	probeIAMIdentityRuntimeTx(t, ctx, database, schema, func(transaction *sql.Tx) {
		_, queryErr = queryIAMUserRelations(ctx, transaction, input)
	})
	if queryErr == nil {
		t.Fatal("IAM user relation mutation unexpectedly succeeded")
	}
	return queryErr
}

func readIAMUserRelationsState(
	t *testing.T, ctx context.Context, database *sql.DB, schema, userID string,
) iamUserRelationsState {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	defer transaction.Rollback()
	var state iamUserRelationsState
	err := transaction.QueryRowContext(ctx, `
SELECT principal.revision,principal.authorization_version,
       COALESCE((SELECT pg_catalog.string_agg(organization_id::text,',' ORDER BY organization_id)
           FROM iam_organization_memberships WHERE user_id=principal.id AND state='active'),''),
       COALESCE((SELECT pg_catalog.string_agg(role_id::text,',' ORDER BY role_id)
           FROM iam_user_role_assignments WHERE user_id=principal.id AND state='active'),'')
FROM iam_users AS principal WHERE principal.id=$1`, userID).Scan(
		&state.revision, &state.authorizationVersion,
		&state.organizations, &state.roles,
	)
	if err != nil {
		t.Fatalf("read IAM user relation state: %v", err)
	}
	return state
}

func countIAMUserRelationReceipt(
	t *testing.T, ctx context.Context, database *sql.DB, schema, operationID string,
) int {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	defer transaction.Rollback()
	var count int
	if err := transaction.QueryRowContext(ctx, `
SELECT pg_catalog.count(*) FROM iam_authorization_operation_receipts
WHERE operation_id=$1`, operationID).Scan(&count); err != nil {
		t.Fatalf("count IAM user relation receipt: %v", err)
	}
	return count
}
