package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	"github.com/google/uuid"
)

type emptyResultRows struct{}

func (emptyResultRows) Next() bool        { return false }
func (emptyResultRows) Scan(...any) error { return nil }
func (emptyResultRows) Err() error        { return nil }
func (emptyResultRows) Close()            {}

type totalProbeRow struct{ total int64 }

func (probe totalProbeRow) Scan(targets ...any) error {
	if len(targets) != 18 {
		return errors.New("unexpected audit probe projection")
	}
	if err := scanAuditFixture(targets[:17]); err != nil {
		return err
	}
	*targets[17].(*int64) = probe.total
	return nil
}

func scanAuditFixture(targets []any) error {
	if len(targets) != 17 {
		return errors.New("unexpected audit entry projection")
	}
	entry := auditFixtureEntry()
	changes, err := encodeFieldChanges(entry.FieldChanges)
	if err != nil {
		return err
	}
	*targets[0].(*uuid.UUID) = uuidArgument(entry.ID)
	*targets[1].(*uuid.UUID) = uuidArgument(entry.OwnerAuditID)
	*targets[2].(*uuid.UUID) = uuidArgument(entry.ActorID)
	*targets[5].(*string), *targets[6].(*string) = entry.OwnerModule, entry.Operation
	*targets[7].(*string), *targets[8].(*string) = entry.EntityType, entry.EntityID
	*targets[10].(*domain.Action) = entry.Action
	*targets[11].(*int64) = entry.AuthorizationVersion
	*targets[13].(*[]byte) = changes
	*targets[14].(*[]byte) = append([]byte(nil), entry.EventSHA256[:]...)
	*targets[15].(*time.Time), *targets[16].(*time.Time) = entry.OccurredAt, entry.RecordedAt
	return nil
}

func auditFixtureEntry() domain.Entry {
	before, after := sha256.Sum256([]byte("old")), sha256.Sum256([]byte("new"))
	now := time.Date(2026, 8, 2, 13, 0, 0, 0, time.UTC)
	return domain.Entry{
		ID:           domain.ID(uuid.MustParse("019c0000-0000-7000-8000-000000000601")),
		OwnerAuditID: domain.ID(uuid.MustParse("019c0000-0000-7000-8000-000000000602")),
		ActorID:      domain.ID(uuid.MustParse("019c0000-0000-7000-8000-000000000603")),
		OwnerModule:  "forms", Operation: "forms.record.update",
		EntityType: "form_record", EntityID: "record-1", Action: domain.ActionUpdate,
		AuthorizationVersion: 3, EventSHA256: sha256.Sum256([]byte("event")),
		OccurredAt: now, RecordedAt: now,
		FieldChanges: []domain.FieldChange{{Field: "password",
			Classification: domain.ClassificationSecret,
			BeforeSHA256:   before, AfterSHA256: after}},
	}
}

func TestListPreservesFilteredTotalForAnEmptyDeepPage(t *testing.T) {
	queryCalls, probeCalls := 0, 0
	store := &Store{scopeAccess: application.ScopeSnapshot,
		usable: func(context.Context) bool { return true },
		query: func(_ context.Context, _ string, arguments ...any) (rows, error) {
			queryCalls++
			if arguments[11] != 2 || arguments[12] != 1 {
				t.Fatalf("page query = %v/%v", arguments[11], arguments[12])
			}
			return emptyResultRows{}, nil
		},
		queryRow: func(_ context.Context, _ string, arguments ...any) row {
			probeCalls++
			if arguments[11] != 1 || arguments[12] != 1 {
				t.Fatalf("total probe = %v/%v", arguments[11], arguments[12])
			}
			return totalProbeRow{total: 1}
		},
		notFound: func(err error) bool { return errors.Is(err, sql.ErrNoRows) },
	}
	page, err := store.List(context.Background(), domain.Filter{Page: 2, Limit: 1})
	if err != nil || page.Total != 1 || page.TotalPages != 1 || len(page.Entries) != 0 ||
		queryCalls != 1 || probeCalls != 1 {
		t.Fatalf("page=%+v query=%d probe=%d err=%v", page, queryCalls, probeCalls, err)
	}
}
