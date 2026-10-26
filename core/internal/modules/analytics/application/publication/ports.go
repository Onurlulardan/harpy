package publication

import (
	"context"
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
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
	Submit(context.Context, TransitionMutation) (Snapshot, bool, error)
	Approve(context.Context, TransitionMutation) (Snapshot, bool, error)
	Activate(context.Context, ActivateMutation) (Snapshot, bool, error)
	Get(context.Context, uuid.UUID) (Snapshot, error)
	List(context.Context, int, *uuid.UUID) (Page, error)
	Revoke(context.Context, TransitionMutation) (Snapshot, bool, error)
}
type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, ScopeAccess) (Store, ScopeIdentity, error)
}
type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, Permission) error
	FreshStepUp(context.Context, platformtx.TenantScope, time.Time) (time.Time, error)
}
type ResourceVerifier interface {
	VerifyImmutable(context.Context, platformtx.TenantScope, ResourceKind,
		uuid.UUID) ([32]byte, error)
}
type IDSource interface{ NewUUIDv7() (uuid.UUID, error) }
type RandomSource interface{ Read([]byte) error }
type Clock interface{ Now() time.Time }

type TransitionEffect struct {
	OperationID, ActorID, TenantID uuid.UUID
	ReleaseID, ResourceVersionID   uuid.UUID
	ResourceKind                   ResourceKind
	Audience                       analyticspublication.Audience
	Generation                     uint64
	OccurredAt                     time.Time
}

type Effects interface {
	Activated(context.Context, platformtx.TenantScope, TransitionEffect) error
	Revoked(context.Context, platformtx.TenantScope, TransitionEffect) error
}

type PublicDirectory interface {
	Route(context.Context, [16]byte) (PublicRoute, error)
}
type PublicScopeCoordinator interface {
	InPublicationSnapshot(context.Context, PublicRoute,
		func(context.Context, platformtx.TenantScope) error) error
}
type PublicExecutor interface {
	Execute(context.Context, platformtx.TenantScope, PublicRoute, []byte) (PublicResult, error)
}
type PublicLimiter interface {
	Admit(context.Context, PublicAdmission) (PublicLease, error)
}
type PublicAdmission struct {
	TenantID, ReleaseID uuid.UUID
	Generation          uint64
	TrustedClientIP     string
}
type PublicLease interface{ Close(context.Context) error }

func permissionForAudience(audience analyticspublication.Audience, approval bool) (Permission, bool) {
	switch {
	case audience == analyticspublication.AudiencePublic && approval:
		return PermissionApprovePublic, true
	case audience == analyticspublication.AudiencePublic:
		return PermissionPublishPublic, true
	case audience == analyticspublication.AudienceCapability && approval:
		return PermissionApproveCapability, true
	case audience == analyticspublication.AudienceCapability:
		return PermissionPublishCapability, true
	case audience == analyticspublication.AudienceTenant:
		return PermissionRead, true
	default:
		return "", false
	}
}
