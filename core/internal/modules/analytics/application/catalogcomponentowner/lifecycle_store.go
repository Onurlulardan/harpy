package catalogcomponentowner

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ResourceBinding struct {
	ResourceID, VersionID, HeadID uuid.UUID
	Lifecycle                     string
	ResourceReferenceEpoch        int64
}

type PurgeSnapshot struct {
	ResourceBinding
	Blockers []componentowner.Blocker
}

type PurgeAuthorityBinding struct {
	OperationID, ComponentID uuid.UUID
	ReferenceEpoch           int64
	AuthoritySHA256          string
	ResourceID               uuid.UUID
	VersionID                uuid.UUID
	HeadID                   uuid.UUID
	RecordedAt               time.Time
}

type LifecycleStore interface {
	Binding(
		context.Context,
		componentowner.Component,
		uuid.UUID,
	) (ResourceBinding, error)
	Preflight(
		context.Context,
		componentowner.Component,
		uuid.UUID,
		uint64,
	) (PurgeSnapshot, error)
	PersistPurgeAuthority(
		context.Context,
		componentowner.Component,
		uuid.UUID,
		PurgeAuthorityBinding,
	) (PurgeAuthorityBinding, error)
	LoadPurgeAuthority(
		context.Context,
		componentowner.Component,
		uuid.UUID,
		uuid.UUID,
	) (PurgeAuthorityBinding, error)
}

type LifecycleStoreBinder interface {
	Bind(
		context.Context,
		platformtx.TenantScope,
		int64,
	) (LifecycleStore, error)
}
