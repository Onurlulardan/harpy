//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func assertIAMProfilePublishRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	fixture iamProfileCommandFixture,
	marker byte,
) {
	t.Helper()
	operationID := profileCommandOperationID(marker, 1)
	requestSHA256 := [32]byte{marker, 1}
	definitionSHA256 := fixture.published.DefinitionSHA256()
	result, err := runIAMProfilePublish(
		ctx, database, tenant.schema, fixture.published,
		uuid.MustParse(iamUserProfileSchemaID), operationID,
		requestSHA256, definitionSHA256,
	)
	if err != nil || result.id != fixture.published.ID() ||
		result.revision != 3 || result.clock != 1 || result.replayed {
		t.Fatalf("profile publish %s = %+v, %v", tenant.schema, result, err)
	}
	replay, err := runIAMProfilePublish(
		ctx, database, tenant.schema, fixture.published,
		uuid.MustParse(iamUserProfileSchemaID), operationID,
		requestSHA256, definitionSHA256,
	)
	if err != nil || !replay.replayed || replay.id != result.id ||
		replay.revision != result.revision || replay.clock != result.clock {
		t.Fatalf("profile publish replay %s = %+v, %v", tenant.schema, replay, err)
	}
	assertIAMProfileFailedPublishRollsBack(
		t, ctx, database, tenant, fixture, marker,
	)
}

func assertIAMProfileFailedPublishRollsBack(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	fixture iamProfileCommandFixture,
	marker byte,
) {
	t.Helper()
	activeField := newIAMProfileCommandField(t, profileapp.FieldActive,
		len(fixture.published.DefinitionJSON())+100)
	next := newIAMProfileCommandSchema(t, iamProfileCommandSchemaV3ID, 3, activeField)
	wrongHash := next.DefinitionSHA256()
	wrongHash[0] ^= 0xff
	_, err := runIAMProfilePublish(
		ctx, database, tenant.schema, next, fixture.published.ID(),
		profileCommandOperationID(marker, 2), [32]byte{marker, 2}, wrongHash,
	)
	if err == nil {
		t.Fatalf("profile publish %s accepted a wrong definition hash", tenant.schema)
	}
	assertIAMSQLState(t, err, "23514")
	var clock int64
	mustQueryIAMTenantValue(t, ctx, database, tenant.schema,
		"SELECT clock_value FROM iam_authorization_clock WHERE singleton", &clock)
	if clock != 1 {
		t.Fatalf("failed profile publish advanced %s clock to %d", tenant.schema, clock)
	}
}

func profileCommandOperationID(marker, operation byte) uuid.UUID {
	value := [16]byte{0xd3, marker, operation}
	value[6], value[8] = 0x40, 0x80
	return uuid.UUID(value)
}
