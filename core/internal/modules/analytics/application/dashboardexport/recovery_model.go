package dashboardexport

import (
	"time"

	"github.com/google/uuid"
)

type UnknownCandidate struct {
	TenantID, ExportID, JobID, AttemptID, LeaseOwnerID, EffectID uuid.UUID
	TenantSchema                                                 string
	AttemptGeneration                                            int64
	EffectRequestSHA256                                          [32]byte
	UnknownAt                                                    time.Time
}

type UnknownClaimDisposition string

const (
	UnknownClaimed  UnknownClaimDisposition = "claimed"
	UnknownBusy     UnknownClaimDisposition = "busy"
	UnknownResolved UnknownClaimDisposition = "resolved"
)

type UnknownCleanupClaim struct {
	Candidate             UnknownCandidate
	Disposition           UnknownClaimDisposition
	CleanupGeneration     int64
	CleanupLeaseOwnerID   uuid.UUID
	CleanupLeaseUntil     time.Time
	Draft                 *ArtifactDraft
	CancellationRequested bool
	ExpiresAt             time.Time
}

type UnknownCleanupLease struct {
	LeaseUntil            time.Time
	CancellationRequested bool
}

type UnknownResolutionKind string

const (
	UnknownResolutionDeleted UnknownResolutionKind = "deleted"
	UnknownResolutionAbsent  UnknownResolutionKind = "absent"
)

type UnknownResolution struct {
	ReceiptID  uuid.UUID
	Kind       UnknownResolutionKind
	SHA256     [32]byte
	Evidence   UnknownResolutionEvidence
	ResolvedAt time.Time
}

type UnknownResolutionEvidence struct {
	ObjectKey, ObjectVersion                            string
	ReconcileProviderRequestID, DeleteProviderRequestID string
	ReconcileEvidenceSHA256, DeleteResponseSHA256       [32]byte
	AbsenceSHA256                                       [32]byte
	UploadedAt, ReconciledAt, DeletedAt                 time.Time
}

type ExpiredCandidate struct {
	TenantID, ExportID, JobID, AttemptID, LeaseOwnerID uuid.UUID
	TenantSchema                                       string
	AttemptGeneration                                  int64
	EffectStarted                                      bool
	EffectID                                           uuid.UUID
	EffectRequestSHA256                                [32]byte
	CancellationPending                                bool
	LeaseExpiredAt                                     time.Time
}

type ExpiredFinalization struct {
	State       RenderExecutionState
	FinalizedAt time.Time
}

type ExpiryControl struct {
	State           State
	CleanupRequired bool
}
