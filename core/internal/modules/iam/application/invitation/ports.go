package invitation

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type EmailEvidence struct {
	NormalizedValue string
	SHA256          [32]byte
}

type EmailNormalizer interface {
	NormalizeEmail(string) (EmailEvidence, error)
}

type Store interface {
	LookupRole(context.Context, platformtx.TenantScope, RoleLookup) (RoleTarget, error)
	LookupCredential(
		context.Context, platformtx.CredentialTenantScope, CredentialLookup,
	) (CredentialRecord, error)
	Create(context.Context, platformtx.TenantScope, CreateCommand) (Ref, error)
	Read(context.Context, platformtx.TenantScope, ReadQuery) (Invitation, error)
	List(context.Context, platformtx.TenantScope, ListQuery) (Page, error)
	Revoke(context.Context, platformtx.TenantScope, RevokeCommand) (Ref, error)
	Expire(context.Context, platformtx.TenantScope, ExpireCommand) (Ref, error)
	Accept(
		context.Context,
		platformtx.CredentialTenantScope,
		AcceptCommand,
	) (AcceptResult, error)
}

type Effects interface {
	WriteTenantChange(
		context.Context, platformtx.TenantScope, AuditChange,
	) error
	WriteCredentialChange(
		context.Context, platformtx.CredentialTenantScope, AuditChange,
	) error
}

type AuditStore interface {
	PersistTenant(context.Context, platformtx.TenantScope, AuditChange) error
	PersistCredential(
		context.Context, platformtx.CredentialTenantScope, AuditChange,
	) error
}
