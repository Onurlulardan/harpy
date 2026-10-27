package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/fullrollout"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fullRolloutRow struct{ value string }

func (row fullRolloutRow) Scan(targets ...any) error {
	reflect.ValueOf(targets[0]).Elem().SetString(row.value)
	return nil
}

type fullRolloutDatabase struct{ arguments []any }

func (database *fullRolloutDatabase) QueryRow(
	_ context.Context, _ string, arguments ...any,
) pgx.Row {
	database.arguments = arguments
	return fullRolloutRow{value: arguments[3].(string)}
}

func TestRecordFullRolloutBindsProtectedReceiptInputs(t *testing.T) {
	database := &fullRolloutDatabase{}
	recorder, err := NewFullRolloutRecorder(database)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("final acceptance"))
	environment := uuid.MustParse("40000000-0000-4000-8000-000000000001")
	tenant := uuid.MustParse("40000000-0000-4000-8000-000000000002")
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	value := fullrollout.PersistedCompletion{
		ReceiptID: environment, EnvironmentID: environment, RecordedAt: now,
		Evidence: fullrollout.CompletionEvidence{
			Steps:            []fullrollout.StepEvidence{{ScopeID: environment}, {ScopeID: tenant}},
			ShadowReadSHA256: digest, LogicalDeltaSHA256: digest,
			ReconcileSHA256: digest, ProductionSnapshotSHA256: digest,
			V1RemovalSHA256: digest, BackupRestoreSHA256: digest,
			FrontendContractSHA256:      digest,
			RollbackWindowClosureSHA256: digest,
			LastLegacyHitAt:             now.Add(-30 * 24 * time.Hour),
			RollbackWindowClosedAt:      now.Add(-time.Hour),
		},
	}
	persisted, err := recorder.Record(context.Background(), value, digest)
	if err != nil || persisted != digest {
		t.Fatalf("persisted digest = %x, %v", persisted, err)
	}
	if len(database.arguments) != 16 || database.arguments[0] != environment ||
		database.arguments[1] != environment || database.arguments[2] != tenant ||
		database.arguments[3] != hex.EncodeToString(digest[:]) ||
		database.arguments[12] != hex.EncodeToString(digest[:]) ||
		database.arguments[15] != now {
		t.Fatalf("full rollout bindings = %#v", database.arguments)
	}
}
