package postgres

import (
	"context"
	"crypto/sha256"
	"reflect"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type groupRehearsalRow struct{ value string }

func (row groupRehearsalRow) Scan(targets ...any) error {
	reflect.ValueOf(targets[0]).Elem().SetString(row.value)
	return nil
}

type groupRehearsalDatabase struct{ arguments []any }

func (database *groupRehearsalDatabase) QueryRow(
	_ context.Context, _ string, arguments ...any,
) pgx.Row {
	database.arguments = arguments
	return groupRehearsalRow{value: arguments[17].(string)}
}

func TestRecordGroupRehearsalBindsEveryOwnerIdentity(t *testing.T) {
	database := &groupRehearsalDatabase{}
	recorder, err := NewGroupRehearsalRecorder(database)
	if err != nil {
		t.Fatal(err)
	}
	value := groupRehearsalFixture()
	if _, err := grouprehearsal.Record(context.Background(), recorder, value); err != nil {
		t.Fatal(err)
	}
	if len(database.arguments) != 19 || database.arguments[0] != value.ReceiptID ||
		database.arguments[1] != value.Evidence.TenantID ||
		database.arguments[3] != value.Evidence.Members[0].OperationID ||
		database.arguments[10] != value.Evidence.Members[1].OperationID {
		t.Fatalf("group rehearsal arguments = %#v", database.arguments)
	}
}

func groupRehearsalFixture() grouprehearsal.PersistedRehearsal {
	digest := sha256.Sum256([]byte("owner-evidence"))
	identity := func(value int) uuid.UUID {
		return uuid.MustParse("20000000-0000-4000-8000-" +
			[]string{"000000000001", "000000000002", "000000000003",
				"000000000004", "000000000005", "000000000006"}[value])
	}
	member := func(domain string, offset int, count uint64) grouprehearsal.MemberEvidence {
		return grouprehearsal.MemberEvidence{Domain: domain,
			OperationID: identity(offset), RunID: identity(offset + 1),
			RollbackReceiptID: identity(offset + 2), SourceCount: count, TargetCount: count,
			AcceptanceSHA256: digest, RollbackSHA256: digest,
			WriterBefore: grouprehearsal.WriterV1, WriterAfter: grouprehearsal.WriterV1}
	}
	return grouprehearsal.PersistedRehearsal{
		ReceiptID: uuid.MustParse("20000000-0000-4000-8000-000000000007"),
		Evidence: grouprehearsal.GroupEvidence{
			TenantID:   uuid.MustParse("20000000-0000-4000-8000-000000000008"),
			Generation: 4, GroupID: grouprehearsal.CanonicalGroupID,
			Members:        []grouprehearsal.MemberEvidence{member("billing", 0, 2), member("forms", 3, 4)},
			QuiesceDrained: true, UsageFlushed: true,
		},
		RecordedAt: time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC),
	}
}
