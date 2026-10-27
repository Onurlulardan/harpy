package jobsv1

const (
	KindPolicyContractVersion       = "job-kind-policy.v1"
	EnqueueContractVersion          = "job-enqueue.v1"
	LeaseContractVersion            = "job-lease.v1"
	AttemptOutcomeContractVersion   = "job-attempt-outcome.v1"
	OperatorReceiptContractVersion  = "job-operator-receipt.v1"
	RetentionReleaseContractVersion = "job-retention-release.v1"

	DefaultMaximumAttempts       = 8
	MinimumMaximumAttempts       = 1
	MaximumMaximumAttempts       = 25
	MaximumClaimBatch            = 50
	LeaseSeconds                 = 60
	HeartbeatSeconds             = 20
	RetryBaseMilliseconds        = 5_000
	RetryCapMilliseconds         = 15 * 60 * 1_000
	SuccessfulRetentionDays      = 90
	DeadRetentionDays            = 365
	PayloadAndErrorRedactionDays = 30
	UnknownOutcomeCode           = "JOB_OUTCOME_UNKNOWN"
	MaximumDomainBytes           = 63
	MaximumKindBytes             = 128
	MaximumMachineCodeBytes      = 64
	MaximumKeyBytes              = 512
	MaximumTicketBytes           = 256
	MaximumReasonBytes           = 2_048
	MaximumReferenceBytes        = 2_048
	MaximumDocumentBytes         = 131_072
)

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateCancelled State = "cancelled"
	StateDead      State = "dead"
)

func (value State) valid() bool {
	switch value {
	case StateQueued, StateRunning, StateSucceeded, StateCancelled, StateDead:
		return true
	default:
		return false
	}
}

type EffectClass string

const (
	EffectDatabaseTransactional EffectClass = "database_transactional"
	EffectExternal              EffectClass = "external"
)

func (value EffectClass) valid() bool {
	return value == EffectDatabaseTransactional || value == EffectExternal
}

type ErrorClass string

const (
	ErrorTransient ErrorClass = "transient"
	ErrorPermanent ErrorClass = "permanent"
	ErrorConflict  ErrorClass = "conflict"
	ErrorCancelled ErrorClass = "cancelled"
)

func (value ErrorClass) valid() bool {
	switch value {
	case ErrorTransient, ErrorPermanent, ErrorConflict, ErrorCancelled:
		return true
	default:
		return false
	}
}

type RetryMode string

const RetryFullJitter RetryMode = "full_jitter"
