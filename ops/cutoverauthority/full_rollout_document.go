package main

type fullRolloutDocument struct {
	ContractVersion          string              `json:"contract_version"`
	ReceiptID                string              `json:"receipt_id"`
	EnvironmentID            string              `json:"environment_id"`
	RecordedAt               string              `json:"recorded_at"`
	Steps                    []fullRolloutStep   `json:"steps"`
	Cohorts                  []fullRolloutCohort `json:"cohorts"`
	Now                      string              `json:"now"`
	LastLegacyHitAt          string              `json:"last_legacy_hit_at"`
	LegacyRouteHits          uint64              `json:"legacy_route_hits"`
	V1Writers                uint64              `json:"v1_writers"`
	V1Consumers              uint64              `json:"v1_consumers"`
	V1QueueMessages          uint64              `json:"v1_queue_messages"`
	V1Credentials            uint64              `json:"v1_credentials"`
	ShadowReadSHA256         string              `json:"shadow_read_sha256"`
	LogicalDeltaSHA256       string              `json:"logical_delta_sha256"`
	ReconcileSHA256          string              `json:"reconcile_sha256"`
	ProductionSnapshotSHA256 string              `json:"production_snapshot_sha256"`
	V1RemovalSHA256          string              `json:"v1_removal_sha256"`
	BackupRestoreSHA256      string              `json:"backup_restore_sha256"`
	FrontendContractSHA256   string              `json:"frontend_contract_sha256"`
	RollbackClosureSHA256    string              `json:"rollback_window_closure_sha256"`
	RollbackWindowClosedAt   string              `json:"rollback_window_closed_at"`
}

type fullRolloutStep struct {
	Scope                  string `json:"scope"`
	Domain                 string `json:"domain"`
	ScopeID                string `json:"scope_id"`
	RequestedBy            string `json:"requested_by"`
	ApprovedBy             string `json:"approved_by"`
	Generation             int64  `json:"generation"`
	State                  string `json:"state"`
	Writer                 string `json:"writer"`
	AcceptanceSHA256       string `json:"acceptance_sha256"`
	ReverseRehearsalSHA256 string `json:"reverse_rehearsal_sha256"`
	SourceCount            uint64 `json:"source_count"`
	TargetCount            uint64 `json:"target_count"`
	SourceSHA256           string `json:"source_sha256"`
	TargetSHA256           string `json:"target_sha256"`
	PendingJobs            uint64 `json:"pending_jobs"`
	PendingOutbox          uint64 `json:"pending_outbox"`
	PendingEffects         uint64 `json:"pending_effects"`
}

type fullRolloutCohort struct {
	Cohort           string             `json:"cohort"`
	StartedAt        string             `json:"started_at"`
	EndedAt          string             `json:"ended_at"`
	EligibleRequests uint64             `json:"eligible_requests"`
	ComparedRequests uint64             `json:"compared_requests"`
	MismatchCount    uint64             `json:"mismatch_count"`
	Metrics          fullRolloutMetrics `json:"metrics"`
}

type fullRolloutMetrics struct {
	V2ErrorBasisPoints             int64  `json:"v2_error_basis_points"`
	BaselineErrorBasisPoints       int64  `json:"baseline_error_basis_points"`
	V2P95Milliseconds              int64  `json:"v2_p95_milliseconds"`
	BaselineP95Milliseconds        int64  `json:"baseline_p95_milliseconds"`
	AuthSuccessBasisPoints         int64  `json:"auth_success_basis_points"`
	BaselineAuthSuccessBasisPoints int64  `json:"baseline_auth_success_basis_points"`
	OldestDueJobLagSeconds         int64  `json:"oldest_due_job_lag_seconds"`
	BaselineJobLagSeconds          int64  `json:"baseline_job_lag_seconds"`
	DLQBasisPoints                 int64  `json:"dlq_basis_points"`
	TenantIsolationIncidents       uint64 `json:"tenant_isolation_incidents"`
	AuthorizationBypasses          uint64 `json:"authorization_bypasses"`
	CredentialLeaks                uint64 `json:"credential_leaks"`
	DataLossIncidents              uint64 `json:"data_loss_incidents"`
	PermanentSecurityJobs          uint64 `json:"permanent_security_jobs"`
}
