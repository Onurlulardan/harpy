//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func assertPublicationWorkerCommitEvidence(
	t *testing.T, ctx context.Context, db publicationDatabase,
	planID, operationID uuid.UUID, expectedRevision int64,
) {
	t.Helper()
	var schemaID uuid.UUID
	var revision, receiptRevision int64
	var document, storedSHA, requestSHA []byte
	var command string
	var preconversions, capabilities int
	err := db.admin.QueryRow(ctx, `SELECT usr.profile_schema_version_id,
usr.revision,usr.profile_values,usr.profile_sha256,receipt.command_key,
receipt.result_revision,receipt.request_sha256,
(SELECT pg_catalog.count(*) FROM
 t_profile_publication.iam_profile_backfill_preconversions
 WHERE plan_id=$1 AND resource_id=usr.id),
(SELECT pg_catalog.count(*) FROM
 t_profile_publication.iam_profile_backfill_worker_capabilities)
FROM t_profile_publication.iam_users AS usr
JOIN t_profile_publication.iam_authorization_operation_receipts AS receipt
 ON receipt.operation_id=$2 WHERE usr.id=$3`,
		planID, operationID, publicationActorID).Scan(
		&schemaID, &revision, &document, &storedSHA, &command,
		&receiptRevision, &requestSHA, &preconversions, &capabilities,
	)
	wantRequest := sha256.Sum256(publicationNicknamePatch)
	if err != nil || schemaID != uuid.MustParse(publicationTargetV3ID) ||
		revision != expectedRevision || receiptRevision != expectedRevision ||
		command != "iam.identity.user.update" || preconversions != 0 ||
		capabilities != 0 || !bytes.Equal(requestSHA, wantRequest[:]) {
		t.Fatalf("worker commit evidence=%s/%d/%s/%d/%d/%d: %v",
			schemaID, revision, command, receiptRevision,
			preconversions, capabilities, err)
	}
	var values map[string]json.RawMessage
	if err = json.Unmarshal(document, &values); err != nil {
		t.Fatalf("decode worker commit profile: %v", err)
	}
	canonical, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	wantSHA := sha256.Sum256(canonical)
	var nickname, replacement string
	_ = json.Unmarshal(values[publicationOptionalID], &nickname)
	_ = json.Unmarshal(values[publicationReplacement], &replacement)
	if !bytes.Equal(storedSHA, wantSHA[:]) || nickname != "nick" ||
		replacement == "" || values[publicationSourceField] != nil {
		t.Fatalf("worker commit profile=%s/%q/%q/%x",
			canonical, nickname, replacement, storedSHA)
	}
}

func assertPublicationRaceTenantIsolation(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime, expectedRevision int64,
) {
	t.Helper()
	var beforeRevision int64
	var beforeSHA []byte
	if err := db.admin.QueryRow(ctx, `SELECT revision,profile_sha256
FROM t_profile_publication.iam_users WHERE id=$1`, publicationActorID).
		Scan(&beforeRevision, &beforeSHA); err != nil {
		t.Fatal(err)
	}
	operationID := uuid.MustParse(
		"57000000-0000-4000-8000-000000000002",
	)
	result, err := executePublicationRaceWriter(
		ctx, runtime, uuid.MustParse(publicationRaceOtherTenantID),
		expectedRevision, publicationNicknamePatch, operationID,
	)
	if err == nil || result.id != uuid.Nil || result.revision != 0 ||
		result.replayed || result.attempts != 2 {
		t.Fatalf("cross-tenant profile writer = %#v/%v", result, err)
	}
	var afterRevision int64
	var afterSHA []byte
	var otherReceipts int
	err = db.admin.QueryRow(ctx, `SELECT usr.revision,usr.profile_sha256,
(SELECT pg_catalog.count(*) FROM
 t_profile_publication_other.iam_authorization_operation_receipts
 WHERE operation_id=$2)
FROM t_profile_publication.iam_users AS usr WHERE usr.id=$1`,
		publicationActorID, operationID).Scan(
		&afterRevision, &afterSHA, &otherReceipts,
	)
	if err != nil || beforeRevision != afterRevision ||
		beforeRevision != expectedRevision || !bytes.Equal(beforeSHA, afterSHA) ||
		otherReceipts != 0 {
		t.Fatalf("cross-tenant profile state=%d/%d/%d: %v",
			beforeRevision, afterRevision, otherReceipts, err)
	}
}
