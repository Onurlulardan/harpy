package queryhttp

import (
	"context"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

type Application interface {
	CreateQuery(context.Context, httptransport.Access, queryapp.CreateRequest) (queryapp.Snapshot, error)
	GetQuery(context.Context, httptransport.Access, uuid.UUID) (queryapp.Snapshot, error)
	GetPublishedQuery(context.Context, httptransport.Access, uuid.UUID) (queryapp.PublishedSnapshot, error)
	ListQueries(context.Context, httptransport.Access, int, *uuid.UUID) (queryapp.Page, error)
	UpdateQuery(context.Context, httptransport.Access, queryapp.UpdateRequest) (queryapp.Snapshot, error)
	DeleteQuery(context.Context, httptransport.Access, queryapp.DeleteRequest) (queryapp.Snapshot, error)
	PublishQuery(context.Context, httptransport.Access, queryapp.PublishRequest) (queryapp.Snapshot, error)
	RevalidateQuery(context.Context, httptransport.Access, queryapp.PublishRequest) (queryapp.Snapshot, error)
	ExecuteQuery(context.Context, httptransport.Access, queryapp.ExecuteRequest) (queryapp.ExecutionResult, error)
}
