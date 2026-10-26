package dashboard

import (
	"context"
	"time"

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
	List(context.Context, int, *uuid.UUID) (Page, error)
	ReadDefault(context.Context, string) (Snapshot, error)
	Update(context.Context, UpdateMutation) (Snapshot, bool, error)
	Delete(context.Context, DeleteMutation) (Snapshot, bool, error)
	Publish(context.Context, PublishMutation) (Snapshot, bool, error)
	GrantACL(context.Context, ACLMutation) (ACL, bool, error)
	RevokeACL(context.Context, ACLMutation) (bool, error)
	SetDefault(context.Context, DefaultMutation) (bool, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, ScopeAccess) (Store, ScopeIdentity, error)
}
type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, Permission) error
	AuthorizePrincipal(context.Context, platformtx.TenantScope, string, uuid.UUID, Permission) error
}
type QueryVersionVerifier interface {
	VerifyPublished(context.Context, platformtx.TenantScope, []uuid.UUID) ([32]byte, error)
}
type IDSource interface{ NewUUIDv7() (uuid.UUID, error) }
type Clock interface{ Now() time.Time }

type PublishEffect struct {
	OperationID, ActorID, TenantID  uuid.UUID
	DashboardID, DashboardVersionID uuid.UUID
	RootRevision, VersionNumber     uint64
	ManifestSHA256                  [32]byte
	OccurredAt                      time.Time
}

type Effects interface {
	DashboardPublished(context.Context, platformtx.TenantScope, PublishEffect) error
}
