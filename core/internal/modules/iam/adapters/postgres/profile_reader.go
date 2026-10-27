package postgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const profileStorageOwner = "iam"

type ProfileReader interface {
	HistoricalSchema(
		context.Context,
		uuid.UUID,
	) (profileapp.Schema, error)
	ActiveSchema(
		context.Context,
		profileapp.ResourceKind,
	) (profileapp.Schema, error)
	HistoricalDocument(
		context.Context,
		profileapp.ResourceKind,
		uuid.UUID,
		profileapp.TextNormalizer,
	) (profileapp.Document, error)
	ActiveQueryCatalog(
		context.Context,
		profileapp.ResourceKind,
	) (profileapp.QueryCatalog, error)
}

type profileRow interface {
	Scan(...any) error
}

type profileQueryRow func(context.Context, string, ...any) profileRow
type profileScopeCheck func(context.Context) bool

type profileReader struct {
	queryRow   profileQueryRow
	scopeCheck profileScopeCheck
}

func (reader *profileReader) usable(ctx context.Context) bool {
	return reader != nil && reader.queryRow != nil && reader.scopeCheck != nil &&
		ctx != nil && ctx.Err() == nil && reader.scopeCheck(ctx)
}
