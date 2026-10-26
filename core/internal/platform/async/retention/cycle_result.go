package retention

import "fmt"

const maximumStepRows int32 = 500

type StepCounts struct {
	JobsRedacted        int32
	OutboxRedacted      int32
	DeadLettersRedacted int32
	InboxPurged         int32
	IdempotencyPurged   int32
	JobsPurged          int32
	DeadLettersPurged   int32
}

type CycleResult struct {
	lockAcquired bool
	counts       StepCounts
}

func NewCycleResult(lockAcquired bool, counts StepCounts) (CycleResult, error) {
	values := [...]int32{
		counts.JobsRedacted, counts.OutboxRedacted, counts.DeadLettersRedacted,
		counts.InboxPurged, counts.IdempotencyPurged, counts.JobsPurged,
		counts.DeadLettersPurged,
	}
	var total int64
	for _, value := range values {
		if value < 0 || value > maximumStepRows {
			return CycleResult{}, fmt.Errorf("%w: step count is outside 0..500", ErrResultInvalid)
		}
		total += int64(value)
	}
	if !lockAcquired && total != 0 {
		return CycleResult{}, fmt.Errorf("%w: contended cycle changed rows", ErrResultInvalid)
	}
	return CycleResult{lockAcquired: lockAcquired, counts: counts}, nil
}

func (result CycleResult) LockAcquired() bool { return result.lockAcquired }

func (result CycleResult) Counts() StepCounts { return result.counts }

func (result CycleResult) Worked() bool {
	counts := result.counts
	return counts.JobsRedacted != 0 || counts.OutboxRedacted != 0 ||
		counts.DeadLettersRedacted != 0 || counts.InboxPurged != 0 ||
		counts.IdempotencyPurged != 0 || counts.JobsPurged != 0 ||
		counts.DeadLettersPurged != 0
}
