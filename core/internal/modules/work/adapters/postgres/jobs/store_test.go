package jobs

import (
	"context"
	"errors"
	"strings"
	"testing"

	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type rowFunc func(...any) error

func (scan rowFunc) Scan(destinations ...any) error { return scan(destinations...) }

func TestStoreRecognizesOnlyExactSucceededReceipt(t *testing.T) {
	claim := adapterClaim()
	generation := claim.LeaseGeneration
	store := &Store{database: databaseAccess{
		execute: func(context.Context, string, ...any) (int64, error) { return 0, nil },
		query:   func(context.Context, string, ...any) (rows, error) { return nil, nil },
		queryRow: func(context.Context, string, ...any) row {
			return rowFunc(func(values ...any) error {
				*values[0].(*string) = claim.Kind
				*values[1].(*[]byte) = append([]byte(nil), claim.PayloadSHA256[:]...)
				*values[2].(*int64) = generation
				*values[3].(*string) = "SUCCEEDED"
				return nil
			})
		},
	}}
	duplicate, err := store.BeginJob(context.Background(), claim)
	if err != nil || !duplicate {
		t.Fatalf("BeginJob exact replay = %v, %v", duplicate, err)
	}
	claim.LeaseGeneration++
	if _, err = store.BeginJob(context.Background(), claim); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("generation-confused replay error = %v", err)
	}
}

func TestWorkJobStaticSQLPinsReceiptsLimitsAndDestructiveGuards(t *testing.T) {
	documents := map[string]string{
		"begin": beginJobSQL, "complete": completeJobSQL,
		"recurrence": insertOccurrenceSQL, "reminder": recordReminderSQL,
		"projection": repairProjectionSQL, "purge": executePurgeSQL,
	}
	required := map[string][]string{
		"begin":      {"work_job_receipts", "ON CONFLICT (job_id) DO NOTHING"},
		"complete":   {"state = 'SUCCEEDED'", "lease_generation = $4"},
		"recurrence": {"work_calendar_events", "ON CONFLICT"},
		"reminder":   {"work_reminder_receipts", "ON CONFLICT"},
		"projection": {"work_projection_streams", "state IN ('GAP','REBUILDING')"},
		"purge": {"cooling_until <= $2", "NOT legal_hold", "blocker_count = 0",
			"live_reference_count = 0", "work_id_tombstones"},
	}
	for name, fragments := range required {
		for _, fragment := range fragments {
			if !strings.Contains(documents[name], fragment) {
				t.Fatalf("%s SQL is missing %q", name, fragment)
			}
		}
	}
}

func adapterClaim() jobapp.Claim {
	claim := jobapp.Claim{JobID: mustAdapterID("019b2ee0-1400-7a1a-8d21-5ca5797f2800"),
		TenantID: mustAdapterID("019b2ee0-1401-7a1a-8d21-5ca5797f2801"),
		TargetID: mustAdapterID("019b2ee0-1402-7a1a-8d21-5ca5797f2802"),
		Kind:     "work.reminder.dispatch", PayloadSHA256: [32]byte{1}, LeaseGeneration: 2,
		Now: mustAdapterTime()}
	return claim
}
