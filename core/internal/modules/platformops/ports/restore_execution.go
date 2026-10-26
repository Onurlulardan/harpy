package ports

import (
	"context"
	"encoding/json"
	"regexp"
	"time"
)

var canonicalRestoreUUID = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

type RestoreExecutionIdentity struct {
	OperationID string
	AttemptID   string
	LeaseToken  string
}

func (identity RestoreExecutionIdentity) Valid() bool {
	return canonicalRestoreUUID.MatchString(identity.OperationID) &&
		canonicalRestoreUUID.MatchString(identity.AttemptID) &&
		identity.AttemptID[14] == '7' &&
		canonicalRestoreUUID.MatchString(identity.LeaseToken) &&
		identity.LeaseToken[14] == '7'
}

type RestoreExecutionClaim struct {
	RestoreExecutionIdentity
	AttemptGeneration            int64
	RestoreState                 string
	RecoverySetID                string
	EnvironmentID                string
	TargetDisposition            string
	PlanSHA256                   string
	PlanBytes                    []byte
	SafetyRecoverySetID          string
	SafetyRecoverySetSHA256      string
	WaiverReason                 string
	ProbeSHA256                  string
	DestructiveConfirmed         bool
	ManifestSHA256               string
	ManifestBytes                []byte
	PreviousDatabaseSystemID     string
	ComponentInventory           json.RawMessage
	ControlDatabaseLocatorSHA256 string
	ScratchDatabaseLocatorSHA256 string
	DrainAckSHA256               string
	SwitchAckSHA256              string
	ArchiveReadbackSHA256        string
	NewDatabaseSystemID          string
	ObjectMerkleSHA256           string
	LocatorRemapSHA256           string
	SafetyEvidenceSHA256         string
	DatabaseNow                  time.Time
}

type RestoreApplyEvidence struct {
	RestoreExecutionIdentity
	DrainAckSHA256               string
	ControlDatabaseLocatorSHA256 string
	ScratchDatabaseLocatorSHA256 string
}

type RestoreVerification struct {
	RestoreExecutionIdentity
	ArchiveReadbackSHA256 string
	NewDatabaseSystemID   string
}

type RestoreSwitchEvidence struct {
	RestoreExecutionIdentity
	ObjectMerkleSHA256           string
	LocatorRemapSHA256           string
	SafetyEvidenceSHA256         string
	SwitchAckSHA256              string
	PreviousDatabaseSystemID     string
	NewDatabaseSystemID          string
	ControlDatabaseLocatorSHA256 string
	ScratchDatabaseLocatorSHA256 string
}

type RestoreCompletion struct {
	RestoreExecutionIdentity
	ReceiptSHA256 string
	ReceiptBytes  []byte
}

type RestoreFailure struct {
	RestoreExecutionIdentity
	Code   string
	Fenced bool
}

type RestoreReceiptImport struct {
	OperationID, RecoverySetID, EnvironmentID     string
	PlanSHA256, ManifestSHA256, ReceiptSHA256     string
	ReceiptBytes                                  []byte
	PreviousDatabaseSystemID, NewDatabaseSystemID string
	ObjectMerkleSHA256, LocatorRemapSHA256        string
	SafetyEvidenceSHA256                          string
	DrainAckSHA256, SwitchAckSHA256               string
	ControlDatabaseLocatorSHA256                  string
	ScratchDatabaseLocatorSHA256                  string
	ComponentInventory                            json.RawMessage
}

type RehydrateControlSeed struct {
	TrustSHA256, PlanSHA256, ManifestSHA256 string
	TrustBytes, PlanBytes, ManifestBytes    []byte
}

type RestoreExecutionStore interface {
	SeedRehydrateControl(context.Context, RehydrateControlSeed) (bool, error)
	ClaimRestoreExecution(
		context.Context, RestoreExecutionIdentity, string,
	) (RestoreExecutionClaim, bool, error)
	UpdateRestoreLease(context.Context, RestoreExecutionIdentity) (bool, error)
	StartRestoreDrain(context.Context, RestoreExecutionIdentity) (bool, error)
	ExecuteRestoreApply(context.Context, RestoreApplyEvidence) (bool, error)
	ValidateRestoreExecution(context.Context, RestoreVerification) (bool, error)
	MarkRestoreSwitched(context.Context, RestoreSwitchEvidence) (bool, error)
	CompleteRestoreExecution(context.Context, RestoreCompletion) (bool, error)
	RejectRestoreExecution(context.Context, RestoreFailure) (bool, error)
	ImportRestoreReceipt(context.Context, RestoreReceiptImport) (bool, error)
}
