package lifecycle

import (
	"context"
	"crypto/ed25519"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ScopeAccess uint8

const (
	ScopeCommand ScopeAccess = iota + 1
	ScopeSnapshot
)

type ScopeIdentity struct {
	ActorID, TenantID uuid.UUID
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, ScopeAccess) (Store, ScopeIdentity, error)
}

type Store interface {
	CreateOperation(context.Context, StartMutation) (OperationHandle, error)
	ReadOperation(context.Context, uuid.UUID) (lifecycle.Operation, error)
	ChangeOperation(context.Context, OperationMutation) (lifecycle.Operation, error)
	RecordOwnerReceipt(context.Context, lifecycle.OwnerReceipt) error
	FinalizePurgeBlockerReceipt(context.Context, PurgeBlockerReceipt) error
	CompileOwnerReceipts(context.Context, uuid.UUID) ([]lifecycle.OwnerReceipt, error)
	GetInstallation(context.Context, packagecontract.PackageID) (lifecycle.Installation, error)
	UpdateInstallation(context.Context, InstallationMutation) (lifecycle.Installation, error)
	ListInstallations(context.Context, PageRequest) (InstallationPage, error)
	ResolveDependencyBlockers(context.Context, packagecontract.PackageID) ([]DependencyBlocker, error)
}

type PurgeAuthorityStore interface {
	AuthorityForPurge(context.Context, uuid.UUID) (PurgeAuthority, error)
}

type PurgeApprovalStore interface {
	ApprovePurge(context.Context, PurgeApprovalMutation) (PurgeApprovalReceipt, bool, error)
	ApprovalReceipt(context.Context, uuid.UUID) (PurgeApprovalReceipt, error)
}

type InstallationByIDStore interface {
	InstallationByID(context.Context, uuid.UUID) (lifecycle.Installation, error)
}

type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, Action) error
}

type JobPort interface {
	Dispatch(context.Context, platformtx.TenantScope, Operation) error
	Abort(context.Context, platformtx.TenantScope, uuid.UUID) error
}

type Operation = lifecycle.Operation

type EventPort interface {
	Emit(context.Context, platformtx.TenantScope, InstallationEvent) error
}

type BackupVerifier interface {
	Verify(
		context.Context,
		platformtx.TenantScope,
		string,
		BackupRequirement,
	) (BackupEvidence, error)
}

type KeyResolver interface {
	Resolve(keyID string) (ed25519.PublicKey, error)
}

type IDSource interface{ New() (uuid.UUID, error) }
type Clock interface{ Now() time.Time }

type BackupGeneration struct {
	Owner      string
	Generation int64
}

type BackupRequirement struct {
	TenantID, PackageID, CatalogOperationID string
	OwnerGenerations                        []BackupGeneration
	Now, MinimumExpiry                      time.Time
}

type BackupEvidence struct {
	ReceiptID, TenantID, PackageID, CatalogOperationID string
	State, SHA256, VerifierVersion                     string
	RestoreVerifiedAt, ExpiresAt                       time.Time
}

type PurgeBlockerReceipt struct {
	ReceiptID, OperationID, ComponentID string
	ReferenceEpoch                      uint64
	CatalogVersion, ZeroDigest          string
	CompletedAt                         time.Time
}
