package v1bridge

import (
	"context"
	"errors"
	"testing"

	bridgeapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1bridge"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type rowFake struct {
	values []any
}

func (row rowFake) Scan(destinations ...any) error {
	if len(destinations) != len(row.values) {
		return errors.New("scan width")
	}
	for index, value := range row.values {
		switch target := destinations[index].(type) {
		case *string:
			*target = value.(string)
		case *int64:
			*target = value.(int64)
		default:
			return errors.New("scan type")
		}
	}
	return nil
}

type databaseFake struct {
	rows  []rowFake
	index int
}

func (database *databaseFake) QueryRow(context.Context, string,
	...any,
) platformpostgres.Row {
	row := database.rows[database.index]
	database.index++
	return row
}

func (*databaseFake) Execute(context.Context, string,
	...any,
) (platformpostgres.CommandResult, error) {
	return platformpostgres.CommandResult{}, errors.New("unused")
}

type enqueuerFake struct{}

func (enqueuerFake) EnqueueNotificationV1Bridge(context.Context, bridgeapp.Job) error {
	return nil
}

func TestStoreLocksProducerAndNotificationAuthorityInCanonicalOrder(t *testing.T) {
	database := &databaseFake{rows: []rowFake{
		{values: []any{"v1", "v1", int64(5)}},
		{values: []any{"v2", "v2", int64(4)}},
		{values: []any{"v1", "writable", int64(5)}},
	}}
	store := &Store{database: database, enqueuer: enqueuerFake{}, tenantID: uuid.New()}
	authority, err := store.LockAuthority(t.Context(), "workflow")
	if err != nil || authority != (bridgeapp.Authority{
		ProducerWriter: bridgeapp.WriterV2, ProducerState: "v2", ProducerGeneration: 4,
		NotificationWriter: bridgeapp.WriterV1, NotificationState: "v1",
		NotificationGeneration: 5,
	}) || database.index != 3 {
		t.Fatalf("authority=%+v calls=%d err=%v", authority, database.index, err)
	}
}

func TestStoreRejectsLocalAndPlatformWriterMismatch(t *testing.T) {
	database := &databaseFake{rows: []rowFake{
		{values: []any{"v1", "v1", int64(5)}},
		{values: []any{"v2", "v2", int64(4)}},
		{values: []any{"v2", "writable", int64(5)}},
	}}
	store := &Store{database: database, enqueuer: enqueuerFake{}, tenantID: uuid.New()}
	if _, err := store.LockAuthority(t.Context(), "workflow"); !errors.Is(err, bridgeapp.ErrAuthority) {
		t.Fatalf("authority error=%v", err)
	}
}
