package grouprehearsal

type Metrics struct {
	V2ErrorBasisPoints             int64
	BaselineErrorBasisPoints       int64
	V2P95Milliseconds              int64
	BaselineP95Milliseconds        int64
	AuthSuccessBasisPoints         int64
	BaselineAuthSuccessBasisPoints int64
	OldestDueJobLagSeconds         int64
	BaselineJobLagSeconds          int64
	DLQBasisPoints                 int64
	TenantIsolationIncidents       uint64
	AuthorizationBypasses          uint64
	CredentialLeaks                uint64
	DataLossIncidents              uint64
	PermanentSecurityJobs          uint64
}

func (value Metrics) Validate() error {
	for _, metric := range []int64{value.V2ErrorBasisPoints, value.BaselineErrorBasisPoints,
		value.V2P95Milliseconds, value.BaselineP95Milliseconds,
		value.AuthSuccessBasisPoints, value.BaselineAuthSuccessBasisPoints,
		value.OldestDueJobLagSeconds, value.BaselineJobLagSeconds, value.DLQBasisPoints} {
		if metric < 0 {
			return ErrInvalid
		}
	}
	if value.V2ErrorBasisPoints > 10_000 || value.BaselineErrorBasisPoints > 10_000 ||
		value.AuthSuccessBasisPoints > 10_000 || value.BaselineAuthSuccessBasisPoints > 10_000 ||
		value.DLQBasisPoints > 10_000 {
		return ErrInvalid
	}
	return nil
}

func (value Metrics) RequiresFence() (bool, error) {
	if err := value.Validate(); err != nil {
		return false, err
	}
	security := value.TenantIsolationIncidents+value.AuthorizationBypasses+
		value.CredentialLeaks+value.DataLossIncidents+value.PermanentSecurityJobs > 0
	errorFence := value.V2ErrorBasisPoints > 50 ||
		value.V2ErrorBasisPoints > 2*value.BaselineErrorBasisPoints &&
			value.V2ErrorBasisPoints-value.BaselineErrorBasisPoints > 20
	latencyFence := value.BaselineP95Milliseconds > 0 &&
		value.V2P95Milliseconds*100 > value.BaselineP95Milliseconds*120
	authFence := value.BaselineAuthSuccessBasisPoints-value.AuthSuccessBasisPoints > 20
	jobFence := value.OldestDueJobLagSeconds > 60 || value.BaselineJobLagSeconds > 0 &&
		value.OldestDueJobLagSeconds > 2*value.BaselineJobLagSeconds
	return security || errorFence || latencyFence || authFence || jobFence ||
		value.DLQBasisPoints > 10, nil
}
