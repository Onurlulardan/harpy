package restoreprogressv1

import "time"

const (
	Schema               = "restore-progress.v1"
	MaximumDocumentBytes = 64 * 1024
	StateReadyToStop     = "ready_to_stop"
	StateRestoring       = "restoring"
	StateReadyToSwitch   = "ready_to_switch"
	StateCompleted       = "completed"
	StateFailed          = "failed"
)

type Payload struct {
	Schema                        string `json:"schema"`
	EnvironmentID                 string `json:"environment_id"`
	OperationID                   string `json:"operation_id"`
	AttemptID                     string `json:"attempt_id"`
	PlanSHA256                    string `json:"plan_sha256"`
	State                         string `json:"state"`
	PreviousDatabaseSystemID      string `json:"previous_database_system_id"`
	NewDatabaseSystemID           string `json:"new_database_system_id,omitempty"`
	PreviousDatabaseLocatorSHA256 string `json:"previous_database_locator_sha256"`
	NewDatabaseLocatorSHA256      string `json:"new_database_locator_sha256"`
	DrainAckSHA256                string `json:"drain_ack_sha256,omitempty"`
	SwitchAckSHA256               string `json:"switch_ack_sha256,omitempty"`
	FailureCode                   string `json:"failure_code,omitempty"`
	OccurredAt                    string `json:"occurred_at"`
}

type Input struct {
	EnvironmentID, OperationID    string
	AttemptID, PlanSHA256         string
	State                         string
	PreviousDatabaseSystemID      string
	NewDatabaseSystemID           string
	PreviousDatabaseLocatorSHA256 string
	NewDatabaseLocatorSHA256      string
	DrainAckSHA256                string
	SwitchAckSHA256               string
	FailureCode                   string
	OccurredAt                    time.Time
}
