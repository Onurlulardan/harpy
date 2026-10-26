package dashboardexport

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ClaimDisposition string

const (
	ClaimExecute   ClaimDisposition = "execute"
	ClaimReconcile ClaimDisposition = "reconcile"
	ClaimTerminal  ClaimDisposition = "terminal"
)

type RenderClaim struct {
	Disposition ClaimDisposition
	Pin
	JobID, LeaseOwner uuid.UUID
	Generation        int64
	Draft             *ArtifactDraft
	State             State
}

type RenderAttemptIdentity struct {
	TenantID, ExportID, JobID, AttemptID, LeaseOwner uuid.UUID
	Generation                                       int64
}

type EffectStart struct {
	Attempt             RenderAttemptIdentity
	EffectID            uuid.UUID
	EffectRequestSHA256 [32]byte
	Draft               ArtifactDraft
	StartedAt           time.Time
}

type ReadyWrite struct {
	Attempt                                 RenderAttemptIdentity
	EffectID, EffectReceiptID               uuid.UUID
	EffectRequestSHA256, EffectResultSHA256 [32]byte
	Receipt                                 UploadReceipt
	CompletedAt                             time.Time
}

type RenderFailure struct {
	Attempt             RenderAttemptIdentity
	EffectID            uuid.UUID
	EffectRequestSHA256 [32]byte
	FailureClass, Code  string
	FailureSHA256       [32]byte
	Retryable           bool
	ProviderReference   string
	FinishedAt          time.Time
}

type RenderUnknown struct {
	Attempt             RenderAttemptIdentity
	EffectID            uuid.UUID
	EffectRequestSHA256 [32]byte
	ObservedAt          time.Time
}

type DeleteClaim struct {
	Pin
	Disposition                            ClaimDisposition
	State                                  State
	Draft                                  *ArtifactDraft
	ObjectVersion                          string
	DeleteActorID, DeleteJobID, LeaseOwner uuid.UUID
	DeleteGeneration                       int64
}

type DeleteReceipt struct {
	ProviderRequestID                   string
	DeleteResponseSHA256, AbsenceSHA256 [32]byte
	VerifiedAbsentAt                    time.Time
}

type RenderResult struct {
	State State
}

type PendingArtifact interface {
	Draft() ArtifactDraft
	Upload(context.Context) (UploadReceipt, error)
	Discard() error
}

type ReconciliationState string

const (
	ReconciliationSealed  ReconciliationState = "sealed"
	ReconciliationMissing ReconciliationState = "missing"
	ReconciliationUnknown ReconciliationState = "unknown"
)

type Reconciliation struct {
	State          ReconciliationState
	Receipt        UploadReceipt
	EvidenceSHA256 [32]byte
	VerifiedAt     time.Time
}

type ArtifactRuntime interface {
	ContentRuntime
	PrepareReserved(
		context.Context, Pin, int64, uuid.UUID, EncryptedSpool, Renderer,
	) (PendingArtifact, error)
	Reconcile(context.Context, ArtifactDraft) (Reconciliation, error)
	Delete(context.Context, ArtifactDraft, string) (DeleteReceipt, error)
}
