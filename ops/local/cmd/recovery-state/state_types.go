package main

type restoreProgress struct {
	Schema      string `json:"schema"`
	OperationID string `json:"operation_id"`
	AttemptID   string `json:"attempt_id"`
	PlanSHA256  string `json:"plan_sha256"`
	State       string `json:"state"`
	UpdatedAt   string `json:"updated_at"`
}

type drainAck struct {
	Schema          string   `json:"schema"`
	OperationID     string   `json:"operation_id"`
	AttemptID       string   `json:"attempt_id"`
	PlanSHA256      string   `json:"plan_sha256"`
	StoppedServices []string `json:"stopped_services"`
	DrainedAt       string   `json:"drained_at"`
}

type emptyProbe struct {
	Schema         string `json:"schema"`
	DatabaseAbsent bool   `json:"database_absent"`
	PolicyAbsent   bool   `json:"policy_absent"`
	ProbedAt       string `json:"probed_at"`
}

type rehydrateRequest struct {
	Schema                string `json:"schema"`
	RecoverySetID         string `json:"recovery_set_id"`
	ExpectedEnvironmentID string `json:"expected_environment_id"`
	TargetDisposition     string `json:"target_disposition"`
	WaiverReason          string `json:"waiver_reason"`
	ProbeSHA256           string `json:"probe_sha256"`
	RecoveryKitSHA256     string `json:"recovery_kit_sha256"`
	DestructiveConfirmed  bool   `json:"destructive_confirmed"`
	RequestedAt           string `json:"requested_at"`
}
