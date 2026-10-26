//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func assertRecoveryFencedWriteError(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime,
) {
	t.Helper()
	var revision int64
	if err := db.admin.QueryRow(ctx, `SELECT revision
FROM t_profile_publication.iam_users WHERE id=$1`, publicationActorID).
		Scan(&revision); err != nil {
		t.Fatal(err)
	}
	operationID := uuid.MustParse("57000000-0000-4000-8000-000000000101")
	result, err := executePublicationRaceWriter(
		ctx, runtime, uuid.MustParse(db.tenant.id), revision,
		publicationNicknamePatch, operationID,
	)
	if !errors.Is(err, profileapp.ErrRecoveryRequired) ||
		result.id != uuid.Nil || result.revision != 0 || result.replayed ||
		result.attempts != 1 {
		t.Fatalf("fenced profile write = %#v/%v", result, err)
	}
	var receipts int
	if err = db.admin.QueryRow(ctx, `SELECT count(*)
FROM t_profile_publication.iam_authorization_operation_receipts
WHERE operation_id=$1`, operationID).Scan(&receipts); err != nil || receipts != 0 {
		t.Fatalf("fenced profile write receipts = %d/%v", receipts, err)
	}
}
