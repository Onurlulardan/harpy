//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

const systemImportSourceRegisterV2SQL = `SELECT source_id,replayed
FROM iam_system_import_source_register_v2(
 pg_catalog.current_schema()::name,$1::uuid,$2::uuid,$3::uuid,$4::text,
 $5::bigint,$6::bigint,$7::bytea,$8::text,$9::text,$10::text,$11::bytea,
 $12::bytea,$13::bigint,$14::bytea,$15::bigint,$16::text,$17::bytea,
 $18::bytea,$19::text,$20::timestamptz)`

type systemImportSourceReplayFixture struct {
	sourceID, receiptID, objectKey, objectVersion string
	header, ciphertext, plaintext, idempotency    []byte
	keyID, provider                               string
	wrapped, nonce                                []byte
	uploadedAt                                    time.Time
}

func assertIAMSystemImportSourceReplay(t *testing.T, ctx context.Context,
	database *sql.DB) {
	t.Helper()
	tenant := iamFoundationTenants[0]
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, tenant.schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	defer transaction.Rollback()
	actorID := iamUserOneID
	const tokenVersion, authorizationVersion int64 = 1, 1
	first := systemImportSourceReplayFixture{
		sourceID:      "019be000-0000-7000-8000-000000000501",
		receiptID:     "019be000-0000-7000-8000-000000000502",
		objectKey:     "iam-system-import-sources/v1/019be000-0000-7000-8000-000000000501/1.fcart",
		objectVersion: "object-version-one", header: bytes.Repeat([]byte{1}, 32),
		ciphertext: bytes.Repeat([]byte{2}, 32), plaintext: bytes.Repeat([]byte{3}, 32),
		idempotency: bytes.Repeat([]byte{4}, 32), keyID: "import-kek-one",
		provider: "provider-one", wrapped: bytes.Repeat([]byte{5}, 48),
		nonce: bytes.Repeat([]byte{6}, 12), uploadedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	storedID, replay, err := registerSystemImportSource(ctx, transaction, actorID,
		tokenVersion, authorizationVersion, first)
	if err != nil || storedID != first.sourceID || replay {
		t.Fatalf("first source=%s replay=%v: %v", storedID, replay, err)
	}
	second := first
	second.sourceID = "019be000-0000-7000-8000-000000000503"
	second.receiptID = "019be000-0000-7000-8000-000000000504"
	second.objectKey = "iam-system-import-sources/v1/019be000-0000-7000-8000-000000000503/1.fcart"
	second.objectVersion = "different-version"
	second.header, second.ciphertext = bytes.Repeat([]byte{7}, 32), bytes.Repeat([]byte{8}, 32)
	second.keyID, second.provider = "different-kek", "different-provider"
	second.wrapped, second.nonce = bytes.Repeat([]byte{9}, 48), bytes.Repeat([]byte{10}, 12)
	storedID, replay, err = registerSystemImportSource(ctx, transaction, actorID,
		tokenVersion, authorizationVersion, second)
	if err != nil || storedID != first.sourceID || !replay {
		t.Fatalf("replayed source=%s replay=%v: %v", storedID, replay, err)
	}
	second.plaintext = bytes.Repeat([]byte{11}, 32)
	_, _, err = registerSystemImportSource(ctx, transaction, actorID,
		tokenVersion, authorizationVersion, second)
	assertIAMSQLState(t, err, "P6102")
}

func registerSystemImportSource(ctx context.Context, transaction *sql.Tx,
	actorID string, tokenVersion, authorizationVersion int64,
	fixture systemImportSourceReplayFixture) (string, bool, error) {
	var sourceID string
	var replay bool
	err := transaction.QueryRowContext(ctx, systemImportSourceRegisterV2SQL,
		fixture.sourceID, fixture.receiptID, actorID, "users", tokenVersion,
		authorizationVersion, fixture.idempotency, "csv", fixture.objectKey,
		fixture.objectVersion, fixture.header, fixture.ciphertext, int64(128),
		fixture.plaintext, int64(64), fixture.keyID, fixture.wrapped,
		fixture.nonce, fixture.provider, fixture.uploadedAt).Scan(&sourceID, &replay)
	return sourceID, replay, err
}
