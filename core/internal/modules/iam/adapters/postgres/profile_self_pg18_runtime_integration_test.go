//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func assertIAMSelfProfileRuntime(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actorID := uuid.MustParse(iamUserThreeID)
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamUserThreeID)
	beforeB := readIAMIdentityActor(t, ctx, database, "t_iam_b", iamUserThreeID)
	read, err := runIAMSelfProfileRead(
		ctx, database, "t_iam_a", actorID,
		actor.tokenVersion, actor.authorizationVersion,
	)
	if err != nil || read.id != actorID || read.email != "self-profile@example.com" ||
		read.status != "active" || read.revision != actor.revision ||
		!json.Valid(read.roles) || !json.Valid(read.organizations) {
		t.Fatalf("self read=%+v error=%v", read, err)
	}
	assertIAMSelfProfileStaleDenied(t, ctx, database, actorID, actor)
	result := assertIAMSelfProfileUpdate(t, ctx, database, actorID, actor)
	assertIAMSelfProfileReplay(t, ctx, database, actorID, actor, result)
	afterB := readIAMIdentityActor(t, ctx, database, "t_iam_b", iamUserThreeID)
	if afterB != beforeB {
		t.Fatalf("tenant B actor changed: before=%+v after=%+v", beforeB, afterB)
	}
	assertIAMSelfProfileRollback(t, ctx, database, actorID)
}

func assertIAMSelfProfileUpdate(
	t *testing.T, ctx context.Context, database *sql.DB,
	actorID uuid.UUID, actor iamIdentityActor,
) iamProfileCommandResult {
	t.Helper()
	schemaID := uuid.MustParse(iamUserProfileSchemaID)
	request := sha256.Sum256([]byte("self-profile-update-a"))
	document := []byte(`{}`)
	digest := sha256.Sum256(document)
	result, err := runIAMSelfProfileUpdate(
		ctx, database, "t_iam_a",
		uuid.MustParse("e1000000-0000-4000-8000-000000000001"), request,
		actorID, schemaID, schemaID, document, digest, actor,
	)
	if err != nil || result.id != actorID || result.revision != actor.revision+1 ||
		result.clock < 1 || result.replayed {
		t.Fatalf("self update=%+v error=%v", result, err)
	}
	return result
}

func assertIAMSelfProfileReplay(
	t *testing.T, ctx context.Context, database *sql.DB,
	actorID uuid.UUID, actor iamIdentityActor, first iamProfileCommandResult,
) {
	t.Helper()
	schemaID := uuid.MustParse(iamUserProfileSchemaID)
	operationID := uuid.MustParse("e1000000-0000-4000-8000-000000000001")
	request := sha256.Sum256([]byte("self-profile-update-a"))
	document, digest := []byte(`{}`), sha256.Sum256([]byte(`{}`))
	replay, err := runIAMSelfProfileUpdate(
		ctx, database, "t_iam_a", operationID, request, actorID,
		schemaID, schemaID, document, digest, actor,
	)
	if err != nil || !replay.replayed || replay.id != first.id ||
		replay.revision != first.revision || replay.clock != first.clock {
		t.Fatalf("self replay=%+v error=%v", replay, err)
	}
	different := sha256.Sum256([]byte("self-profile-update-b"))
	_, err = runIAMSelfProfileUpdate(
		ctx, database, "t_iam_a", operationID, different, actorID,
		schemaID, schemaID, document, digest, actor,
	)
	if !iamSelfProfileSQLState(err, "23505") {
		t.Fatalf("self replay payload reuse error=%v", err)
	}
}

func assertIAMSelfProfileStaleDenied(
	t *testing.T, ctx context.Context, database *sql.DB,
	actorID uuid.UUID, actor iamIdentityActor,
) {
	t.Helper()
	_, err := runIAMSelfProfileRead(
		ctx, database, "t_iam_a", actorID,
		actor.tokenVersion+1, actor.authorizationVersion,
	)
	if !iamSelfProfileSQLState(err, "42501") {
		t.Fatalf("stale self read error=%v", err)
	}
}

func iamSelfProfileSQLState(err error, state string) bool {
	var failure *pgconn.PgError
	return errors.As(err, &failure) && failure.Code == state
}
