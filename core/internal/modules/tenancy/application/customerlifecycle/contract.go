package customerlifecycle

import (
	"context"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type CredentialEnvelope struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
}

type AccessRequest struct {
	Credential     CredentialEnvelope
	Capability     string
	ExpectedMode   string
	TargetTenantID uuid.UUID
	RequireStepUp  bool
	StepUpPurpose  [32]byte
}

type AccessDecision struct {
	SubjectID, ControlTenantID, SessionID uuid.UUID
	DeploymentMode, Capability            string
	StepUpAt                              time.Time
}

type AccessVerifier interface {
	Verify(context.Context, AccessRequest) (AccessDecision, error)
}

type DeploymentPolicy interface {
	Snapshot() (tenantapi.DeploymentPolicy, error)
	PrimaryTarget() (tenantapi.PrimaryTenantTarget, error)
}

type ListStoreRequest struct {
	EnvironmentID uuid.UUID
	Limit         int
	After         uuid.UUID
}

type MutationStoreRequest struct {
	EnvironmentID, ControlTenantID, ActorID, TargetTenantID uuid.UUID
	Kind, Name, IdempotencyKey                              string
	ExpectedRevision                                        int64
	RequestSHA256                                           [32]byte
	Now                                                     time.Time
}

type MutationStoreResult struct {
	Tenant  Tenant
	Outcome string
	Replay  bool
}

type AsyncStoreRequest struct {
	EnvironmentID, ControlTenantID, ActorID, TargetTenantID uuid.UUID
	Kind, IdempotencyKey                                    string
	ExpectedRevision                                        int64
	RequestSHA256                                           [32]byte
	Proof                                                   PurgeProof
	Now                                                     time.Time
}

type AsyncStoreResult struct {
	Operation Operation
	Outcome   string
}

type Store interface {
	List(context.Context, ListStoreRequest) (Page, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Tenant, error)
	Mutate(context.Context, MutationStoreRequest) (MutationStoreResult, error)
	Accept(context.Context, AsyncStoreRequest) (AsyncStoreResult, error)
	ReadOperation(context.Context, uuid.UUID, uuid.UUID) (Operation, error)
}

type WorkerStore interface {
	Claim(context.Context, uuid.UUID, time.Time, time.Duration) (ClaimedOperation, error)
	CompleteOffboard(context.Context, ClaimedOperation, time.Time) (Operation, error)
	CompletePurge(context.Context, ClaimedOperation, ExecutionReceipt, time.Time) (Operation, error)
	Fail(context.Context, ClaimedOperation, string, time.Time) error
}

type PurgeReadiness interface {
	Verify(context.Context, uuid.UUID) (PurgeProof, error)
}

type DestructionExecutor interface {
	Purge(context.Context, ClaimedOperation) (ExecutionReceipt, error)
}

type Clock interface{ Now() time.Time }
