package systemtransfercleanup

import (
	"context"
	"regexp"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

var schemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type row interface{ Scan(...any) error }
type database interface {
	QueryRow(context.Context, string, ...any) row
}
type rowNotFound func(error) bool

type Store struct {
	database database
	notFound rowNotFound
	inTenant func(context.Context, uuid.UUID,
		func(context.Context, *tenantStore) error) error
}

type tenantStore struct {
	database      database
	currentSchema func(context.Context) (string, error)
	notFound      rowNotFound
}

func (store *Store) valid() bool {
	return store != nil && store.database != nil && store.notFound != nil &&
		store.inTenant != nil
}

func (store *Store) runInTenant(ctx context.Context, tenantID uuid.UUID,
	execute func(context.Context, *tenantStore) error,
) error {
	if !store.valid() || ctx == nil || ctx.Err() != nil || tenantID == uuid.Nil ||
		execute == nil {
		return app.ErrSourceCleanupUnavailable
	}
	return store.inTenant(ctx, tenantID, execute)
}

func (store *tenantStore) schema(ctx context.Context) (string, error) {
	if store == nil || store.currentSchema == nil {
		return "", app.ErrSourceCleanupUnavailable
	}
	schema, err := store.currentSchema(ctx)
	if err != nil || !schemaPattern.MatchString(schema) {
		return "", app.ErrSourceCleanupUnavailable
	}
	return schema, nil
}

func validTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0
}
