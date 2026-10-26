package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
)

type appendProbeRow struct{ replay bool }

func (probe appendProbeRow) Scan(targets ...any) error {
	if err := scanAuditFixture(targets[:17]); err != nil {
		return err
	}
	*targets[17].(*bool) = probe.replay
	return nil
}

func TestAppendUsesOwnedRoutineAndPreservesReplayIdentity(t *testing.T) {
	entry := auditFixtureEntry()
	queryCalls := 0
	store := &Store{scopeAccess: application.ScopeCommand,
		usable: func(context.Context) bool { return true },
		query: func(context.Context, string, ...any) (rows, error) {
			return emptyResultRows{}, nil
		},
		queryRow: func(_ context.Context, statement string, arguments ...any) row {
			queryCalls++
			if statement != appendSQL || len(arguments) != 17 ||
				arguments[0] != uuidArgument(entry.ID) ||
				arguments[1] != uuidArgument(entry.OwnerAuditID) {
				t.Fatalf("append routine arguments drifted")
			}
			return appendProbeRow{}
		},
	}
	persisted, replay, err := store.Append(context.Background(), application.AppendMutation{
		Entry: entry, RequestSHA256: entry.EventSHA256})
	if err != nil || replay || queryCalls != 1 || persisted.ID != entry.ID ||
		persisted.OwnerAuditID != entry.OwnerAuditID ||
		persisted.EventSHA256 != entry.EventSHA256 || persisted.Validate() != nil {
		t.Fatalf("persisted=%+v replay=%v calls=%d err=%v", persisted, replay, queryCalls, err)
	}
}
