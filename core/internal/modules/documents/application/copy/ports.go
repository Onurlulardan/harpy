package copy

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type SourceReader interface {
	Open(context.Context, ScopeRunner, domain.ID, domain.ID) (Source, error)
}

type DestinationWriter interface {
	CreateAndPut(context.Context, ScopeRunner, DestinationRequest) (Result, error)
}

type ScopeRunner interface {
	InCommand(context.Context, string,
		func(context.Context, platformtx.TenantScope) error) error
	InSnapshot(context.Context, string,
		func(context.Context, platformtx.TenantScope) error) error
}

type LeaseSource interface{ NewUUIDv7() (domain.ID, error) }

type Dependencies struct {
	Source      SourceReader
	Destination DestinationWriter
	Authorizer  foundation.Authorizer
	LeaseIDs    LeaseSource
}
