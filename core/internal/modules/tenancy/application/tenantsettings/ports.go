package tenantsettings

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

// Scope keeps storage adapters behind the Tenancy application boundary while
// preserving the exact platform transaction capability type.
type Scope = platformtx.TenantScope

type RegistryMutation struct {
	Name             *string
	ExpectedRevision int64
	IdempotencyKey   string
	RequestSHA256    [32]byte
}

type Registry interface {
	Read(context.Context, Scope) (RegistrySnapshot, error)
	Update(context.Context, Scope, RegistryMutation) (UpdateResult, error)
}

type ProjectionReader interface {
	Read(context.Context, Scope) (Projection, error)
}

type MailMutation struct {
	SMTP             SMTPMutation
	ExpectedRevision int64
	ResultRevision   int64
	IdempotencyKey   string
	RequestSHA256    [32]byte
}

type TestMailRequest struct {
	TenantID         uuid.UUID
	TenantName       string
	To               string
	ExpectedRevision int64
	IdempotencyKey   string
	RequestSHA256    [32]byte
	RequestedAt      time.Time
}

type MailSettings interface {
	// Replace must be exact replay-safe for one tenant, idempotency key and
	// request digest. An identical replay returns success without repeating the
	// externally observable replacement; a reused key with a different digest
	// must conflict. ExpectedRevision and ResultRevision are part of that fence.
	Replace(context.Context, Scope, MailMutation) error
	EnqueueTest(context.Context, Scope, TestMailRequest) (TestEmailResult, error)
}

type Clock interface{ Now() time.Time }
