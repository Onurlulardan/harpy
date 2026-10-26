package audithttp

import (
	"context"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	"github.com/google/uuid"
)

type Filter = domain.Filter
type Page = domain.Page
type Stats = domain.Stats
type Entry = domain.Entry

type Application interface {
	List(context.Context, Access, Filter, string) (Page, error)
	Stats(context.Context, Access) (Stats, error)
	Get(context.Context, Access, uuid.UUID) (Entry, error)
}

type ErrorSink func(error)
type CorrelationResolver func(*http.Request) string

type Route struct {
	Method, Pattern string
	Handler         http.Handler
}
