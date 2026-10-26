package menu

import (
	"context"
	"time"

	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ScopeAccess uint8

const (
	ScopeCommand ScopeAccess = iota + 1
	ScopeSnapshot
)

type Identity struct{ ActorID, TenantID uuid.UUID }

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, ScopeAccess) (Store, Identity, error)
}

type Store interface {
	Get(context.Context) (Source, error)
	GetActiveConfig(context.Context) (ConfigDocument, error)
	CreateConfig(context.Context, ConfigMutation) (ConfigDocument, error)
	UpdateConfig(context.Context, ConfigMutation) (ConfigDocument, error)
	Update(context.Context, OverrideMutation) (OverrideReceipt, error)
	Replace(context.Context, ReplaceMutation) (Source, error)
}

type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, string) error
}

type TargetAuthorizer interface {
	Allows(context.Context, platformtx.TenantScope,
		packagecontract.MenuKind, packagecontract.ComponentID) (bool, error)
}

type Clock interface{ Now() time.Time }
type IDSource interface{ New() (uuid.UUID, error) }

type Source struct {
	Version        menudomain.Version
	Overrides      []menudomain.Override
	ActivePackages map[string]bool
}

type OverrideMutation struct {
	Override         menudomain.Override
	ReceiptID        uuid.UUID
	ExpectedRevision uint64
	ActorID          uuid.UUID
	IdempotencyKey   string
	RequestSHA256    string
	OccurredAt       time.Time
}

type OverrideReceipt struct {
	Override menudomain.Override
	Revision uint64
	Replayed bool
}

type ReplaceMutation struct {
	VersionID        uuid.UUID
	ExpectedRevision uint64
	Items            []menudomain.Item
	OccurredAt       time.Time
	ContentSHA256    string
}
