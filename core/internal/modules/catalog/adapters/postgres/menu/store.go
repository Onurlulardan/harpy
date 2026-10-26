package menu

import (
	"context"
	"regexp"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	"github.com/google/uuid"
)

const storageOwner = "catalog"

var schemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$`)

type row interface{ Scan(...any) error }
type rows interface {
	Close()
	Err() error
	Next() bool
	Scan(...any) error
}

type store struct {
	tenantID, actorID uuid.UUID
	command           bool
	queryRow          func(context.Context, string, ...any) row
	query             func(context.Context, string, ...any) (rows, error)
	execute           func(context.Context, string, ...any) (int64, error)
	currentSchema     func(context.Context) (string, error)
	notFound          func(error) bool
	allowed           func(context.Context) bool
}

func (storage *store) usable(ctx context.Context, command bool) bool {
	if storage == nil || storage.tenantID == uuid.Nil || storage.actorID == uuid.Nil ||
		storage.queryRow == nil || storage.query == nil || storage.execute == nil ||
		storage.currentSchema == nil || storage.notFound == nil || storage.allowed == nil ||
		ctx == nil || ctx.Err() != nil ||
		!storage.allowed(ctx) || command && !storage.command {
		return false
	}
	schema, err := storage.currentSchema(ctx)
	return err == nil && schemaPattern.MatchString(schema) && schema != storageOwner
}

func unavailable() error { return menuapp.ErrUnavailable }

var _ menuapp.Store = (*store)(nil)
