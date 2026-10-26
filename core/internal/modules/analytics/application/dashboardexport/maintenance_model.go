package dashboardexport

import (
	"time"

	"github.com/google/uuid"
)

const (
	maintenanceDeleteKind = "analytics.dashboard.export.delete"
	maintenanceExpireKind = "analytics.dashboard.export.expire"
)

type MaintenanceUnknownCandidate struct {
	Kind                                             string
	TenantID                                         uuid.UUID
	TenantSchema, AuthorityGroupID                   string
	ExportID, JobID, AttemptID, OriginalLeaseOwnerID uuid.UUID
	EffectID                                         uuid.UUID
	AuthorityGroupGeneration, CutoverGeneration      int64
	AttemptNumber, LeaseGeneration                   int64
	EffectRequestSHA256                              [32]byte
	UnknownAt                                        time.Time
}

type MaintenanceCleanupMode string

const (
	MaintenanceCleanupArtifact  MaintenanceCleanupMode = "artifact"
	MaintenanceCleanupTerminal  MaintenanceCleanupMode = "terminal"
	MaintenanceCleanupCompleted MaintenanceCleanupMode = "completed"
)

type MaintenanceUnknownClaim struct {
	Candidate               MaintenanceUnknownCandidate
	Disposition             UnknownClaimDisposition
	CleanupGeneration       int64
	DeleteAttemptGeneration int64
	CleanupLeaseOwnerID     uuid.UUID
	CleanupLeaseUntil       time.Time
	Mode                    MaintenanceCleanupMode
	TerminalState           State
	Draft                   *ArtifactDraft
	ObjectVersion           string
	StoredEvidence          MaintenanceResolutionEvidence
}

type MaintenanceResolutionEvidence struct {
	ObjectKey, ObjectVersion                            string
	ReconcileProviderRequestID, DeleteProviderRequestID string
	ReconcileEvidenceSHA256, DeleteResponseSHA256       [32]byte
	AbsenceSHA256, LocalStateSHA256                     [32]byte
	UploadedAt, ReconciledAt, DeletedAt, LocalStateAt   time.Time
}

type MaintenanceResolution struct {
	ReceiptID     uuid.UUID
	Kind          UnknownResolutionKind
	SHA256        [32]byte
	TerminalState State
	Evidence      MaintenanceResolutionEvidence
	ResolvedAt    time.Time
}

func validMaintenanceCandidate(candidate MaintenanceUnknownCandidate) bool {
	maximumAttempts := int64(3)
	if candidate.Kind == maintenanceExpireKind {
		maximumAttempts = 8
	} else if candidate.Kind != maintenanceDeleteKind {
		return false
	}
	return candidate.TenantID != uuid.Nil &&
		tenantSchemaPattern.MatchString(candidate.TenantSchema) &&
		candidate.AuthorityGroupID != "" &&
		validUUIDv7(candidate.ExportID) && validUUIDv7(candidate.JobID) &&
		validUUIDv7(candidate.AttemptID) &&
		validUUIDv7(candidate.OriginalLeaseOwnerID) &&
		candidate.EffectID.Version() == uuid.Version(5) &&
		candidate.AuthorityGroupGeneration > 0 && candidate.CutoverGeneration > 0 &&
		candidate.AttemptNumber >= 1 && candidate.AttemptNumber <= maximumAttempts &&
		candidate.LeaseGeneration == candidate.AttemptNumber &&
		candidate.EffectRequestSHA256 != [32]byte{} && !candidate.UnknownAt.IsZero()
}
