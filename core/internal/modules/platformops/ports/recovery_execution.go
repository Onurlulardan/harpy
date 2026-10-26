package ports

import (
	"context"
	"encoding/json"
	"regexp"
	"time"
)

var canonicalExecutionUUID = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

type RecoveryExecutionClaim struct {
	AttemptID                 string
	LeaseToken                string
	AttemptGeneration         int64
	RecoveryState             string
	RecoverySetID             string
	OperationID               string
	EnvironmentID             string
	Sequence                  int64
	PrimaryBindingID          string
	TenantSetSHA256           string
	OwnerBindingID            string
	OwnerBindingGeneration    int64
	DatabaseLineageID         string
	DatabaseSystemID          string
	SigningGeneration         int64
	RecoveryTrustBundleSHA256 string
	ComponentInventory        json.RawMessage
	DatabaseNow               time.Time
}

type RecoveryComponentRecord struct {
	RecoverySetID   string
	AttemptID       string
	LeaseToken      string
	Kind            string
	ObjectKey       string
	ObjectVersion   string
	CipherSHA256    string
	CiphertextBytes int64
}

type RecoveryComponentBatch struct {
	RecoveryExecutionIdentity
	Components json.RawMessage
}

type RecoveryCompletionRecord struct {
	RecoverySetID    string
	AttemptID        string
	LeaseToken       string
	ManifestSHA256   string
	ManifestBytes    []byte
	DatabaseSystemID string
}

type RecoveryExecutionStore interface {
	ClaimNextRecoverySet(context.Context, string, string) (RecoveryExecutionClaim, bool, error)
	UpdateRecoveryLease(context.Context, RecoveryExecutionIdentity) (bool, error)
	AttestDocumentsHold(context.Context, DocumentsHoldAttestation) (bool, error)
	RecordRecoveryComponent(context.Context, RecoveryComponentRecord) (bool, error)
	PutRecoveryComponents(context.Context, RecoveryComponentBatch) (bool, error)
	RejectRecoveryExecution(context.Context, RecoveryExecutionFailure) (bool, error)
	CompleteRecoverySet(context.Context, RecoveryCompletionRecord) (bool, error)
}

type RecoveryExecutionIdentity struct {
	RecoverySetID string
	AttemptID     string
	LeaseToken    string
}

func (identity RecoveryExecutionIdentity) Valid() bool {
	return canonicalExecutionUUID.MatchString(identity.RecoverySetID) &&
		canonicalExecutionUUID.MatchString(identity.AttemptID) &&
		identity.AttemptID[14] == '7' &&
		canonicalExecutionUUID.MatchString(identity.LeaseToken) &&
		identity.LeaseToken[14] == '7'
}

type DocumentsHoldAttestation struct {
	RecoveryExecutionIdentity
	InventorySHA256 string
	InventoryCount  int64
}

type RecoveryExecutionFailure struct {
	RecoveryExecutionIdentity
	Code string
}
