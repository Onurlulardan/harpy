//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func assertIAMProfileBackfillRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	fixture iamProfileCommandFixture,
	marker byte,
) {
	t.Helper()
	var expectedRevision int64
	mustQueryIAMTenantValue(t, ctx, database, tenant.schema,
		"SELECT revision FROM iam_users WHERE id='"+iamUserThreeID+"'",
		&expectedRevision)
	operationID := profileCommandOperationID(marker, 3)
	requestSHA256 := [32]byte{marker, 3}
	result, err := runIAMProfileBackfill(
		ctx, database, tenant.schema, uuid.MustParse(iamUserThreeID),
		uuid.MustParse(iamUserProfileSchemaID), expectedRevision,
		fixture.document, operationID, requestSHA256,
	)
	if err != nil || result.id != uuid.MustParse(iamUserThreeID) ||
		result.revision != expectedRevision+1 || result.clock != 2 || result.replayed {
		t.Fatalf("profile backfill %s = %+v, %v", tenant.schema, result, err)
	}
	replay, err := runIAMProfileBackfill(
		ctx, database, tenant.schema, uuid.MustParse(iamUserThreeID),
		uuid.MustParse(iamUserProfileSchemaID), expectedRevision,
		fixture.document, operationID, requestSHA256,
	)
	if err != nil || !replay.replayed || replay.id != result.id ||
		replay.revision != result.revision || replay.clock != result.clock {
		t.Fatalf("profile backfill replay %s = %+v, %v", tenant.schema, replay, err)
	}
	assertIAMProfileBackfillStoredExactDocument(t, ctx, database, tenant, fixture)
}

func assertIAMProfileBackfillStoredExactDocument(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	fixture iamProfileCommandFixture,
) {
	t.Helper()
	var schemaID string
	var storedSHA []byte
	mustQueryIAMTenantValue(t, ctx, database, tenant.schema, `
SELECT profile_schema_version_id::text, profile_sha256
FROM iam_users WHERE id='`+iamUserThreeID+`'`, &schemaID, &storedSHA)
	wantSHA := fixture.document.SHA256()
	if schemaID != iamProfileCommandSchemaV2ID ||
		string(storedSHA) != string(wantSHA[:]) {
		t.Fatalf("profile backfill %s stored %s/%x", tenant.schema, schemaID, storedSHA)
	}
}
