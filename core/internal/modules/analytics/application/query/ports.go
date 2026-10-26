package query

import (
	"context"
	"time"

	analyticsforms "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/formsreference"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ScopeAccess uint8

const (
	ScopeCommand ScopeAccess = iota + 1
	ScopeSnapshot
)

type ScopeIdentity struct{ ActorID, TenantID uuid.UUID }

type Store interface {
	Create(context.Context, CreateMutation) (Snapshot, bool, error)
	Get(context.Context, uuid.UUID) (Snapshot, error)
	LatestPublished(context.Context, uuid.UUID) (Snapshot, error)
	ResolveQueryID(context.Context, uuid.UUID) (uuid.UUID, error)
	LoadVersion(context.Context, uuid.UUID, uuid.UUID) (Snapshot, error)
	Dependencies(context.Context, uuid.UUID, uuid.UUID) ([]DependencySnapshot, error)
	List(context.Context, int, *uuid.UUID) (Page, error)
	Update(context.Context, UpdateMutation) (Snapshot, bool, error)
	Delete(context.Context, DeleteMutation) (Snapshot, bool, error)
	Publish(context.Context, PublishMutation) (Snapshot, bool, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, ScopeAccess) (Store, ScopeIdentity, error)
}

type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, Permission) error
}

type DependencyResolver interface {
	Dependencies(context.Context, platformtx.TenantScope,
		[]DependencyRequest) ([]DependencySnapshot, error)
}

type ContractPort interface {
	Normalize(context.Context, platformtx.TenantScope, Document, Document,
		Document, []DependencySnapshot) (NormalizedContract, error)
	BindExecution(context.Context, platformtx.TenantScope, Snapshot,
		[]DependencySnapshot, Document) (ExecutionBinding, error)
	Run(context.Context, platformtx.TenantScope, Snapshot,
		[]DependencySnapshot, Document) (ExecutionResult, error)
}

type Cache interface {
	Read(context.Context, CacheRequest) ([]byte, bool, error)
	Write(context.Context, CacheRequest, []byte, time.Time) error
	Invalidate(context.Context, uuid.UUID) error
}

type PublishEffect struct {
	OperationID, ActorID, TenantID uuid.UUID
	QueryID, QueryVersionID        uuid.UUID
	RootRevision, VersionNumber    uint64
	ManifestSHA256                 [32]byte
	OccurredAt                     time.Time
	Revalidated                    bool
}

type Effects interface {
	QueryPublished(context.Context, platformtx.TenantScope, PublishEffect) error
}

type FormsReferenceCoordinator interface {
	BindVersion(context.Context, platformtx.TenantScope,
		analyticsforms.Source, []analyticsforms.Target, time.Time) error
	DecommissionSource(context.Context, platformtx.TenantScope,
		analyticsforms.ID, *analyticsforms.ID, time.Time) error
}

type CacheRequest struct {
	TenantID                      uuid.UUID
	QueryID                       uuid.UUID
	QueryVersionID                uuid.UUID
	Binding                       ExecutionBinding
	QuerySHA256, ParametersSHA256 [32]byte
}

type IDSource interface{ NewUUIDv7() (uuid.UUID, error) }
type Clock interface{ Now() time.Time }
