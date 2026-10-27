package recoveryset

type RestorePlanPayload struct {
	Schema                  string              `json:"schema"`
	RecoverySetID           string              `json:"recovery_set_id"`
	ManifestSHA256          string              `json:"manifest_sha256"`
	Identity                EnvironmentIdentity `json:"identity"`
	TargetDisposition       TargetDisposition   `json:"target_disposition"`
	SafetyRecoverySetID     string              `json:"safety_recovery_set_id,omitempty"`
	SafetyRecoverySetSHA256 string              `json:"safety_recovery_set_sha256,omitempty"`
	WaiverReason            string              `json:"waiver_reason,omitempty"`
	ProbeSHA256             string              `json:"probe_sha256,omitempty"`
	DestructiveConfirmed    bool                `json:"destructive_confirmed"`
}

type RestoreReceiptPayload struct {
	Schema                        string              `json:"schema"`
	OperationID                   string              `json:"operation_id"`
	RecoverySetID                 string              `json:"recovery_set_id"`
	PlanSHA256                    string              `json:"plan_sha256"`
	ManifestSHA256                string              `json:"manifest_sha256"`
	Identity                      EnvironmentIdentity `json:"identity"`
	PreviousDatabaseSystemID      string              `json:"previous_database_system_id,omitempty"`
	NewDatabaseSystemID           string              `json:"new_database_system_id"`
	PreviousDatabaseLocatorSHA256 string              `json:"previous_database_locator_sha256"`
	NewDatabaseLocatorSHA256      string              `json:"new_database_locator_sha256"`
	DrainAckSHA256                string              `json:"drain_ack_sha256"`
	SwitchAckSHA256               string              `json:"switch_ack_sha256"`
	ObjectMerkleSHA256            string              `json:"object_merkle_sha256"`
	LocatorRemapSHA256            string              `json:"locator_remap_sha256"`
	SafetyEvidenceSHA256          string              `json:"safety_evidence_sha256"`
	Seeded                        bool                `json:"seeded"`
	OwnerChanged                  bool                `json:"owner_changed"`
	TenantReplaced                bool                `json:"tenant_replaced"`
	SwitchOutcome                 string              `json:"switch_outcome"`
}

const (
	RestorePlanSchema    = "restore-plan-payload.v1"
	RestoreReceiptSchema = "restore-receipt-payload.v2"
)
