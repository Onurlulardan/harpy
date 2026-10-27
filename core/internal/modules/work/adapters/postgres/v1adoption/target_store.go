package v1adoption

import (
	"context"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type RecordWriter interface {
	Persist(context.Context, adoption.TargetRecord) error
	Revert(context.Context, adoption.ReverseRequest, adoption.Evidence) error
}

type TargetStore struct {
	database DatabaseAccess
	tenantID domain.ID
	writer   RecordWriter
}

func NewTargetStore(database DatabaseAccess, tenantID domain.ID,
	writer RecordWriter,
) (*TargetStore, error) {
	if database.Execute == nil || database.Query == nil || database.QueryRow == nil ||
		database.CurrentSchema == nil || !tenantID.Valid() || writer == nil {
		return nil, ports.ErrInvalid
	}
	return &TargetStore{database: database, tenantID: tenantID, writer: writer}, nil
}

func (store *TargetStore) check(ctx context.Context, scope ports.Scope) error {
	if store == nil || scope == nil || scope.TenantID() != store.tenantID ||
		ctx == nil || ctx.Err() != nil {
		return ports.ErrInvalid
	}
	schema, err := store.database.CurrentSchema(ctx)
	if err != nil || !schemaPattern.MatchString(schema) || schema == "public" {
		return ports.ErrUnavailable
	}
	return nil
}

func (store *TargetStore) executeOne(ctx context.Context, statement string, arguments ...any) error {
	affected, err := store.database.Execute(ctx, statement, arguments...)
	if err != nil {
		return ports.ErrUnavailable
	}
	if affected != 1 {
		return ports.ErrConflict
	}
	return nil
}

var _ adoption.Target = (*TargetStore)(nil)
