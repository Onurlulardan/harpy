package actorbridge

import "time"

type Reconciliation struct {
	accountedCount int64
	latestHitAt    *time.Time
}

func NewReconciliation(accounted int64, latest *time.Time) (Reconciliation, error) {
	if accounted < 0 || (accounted == 0) != (latest == nil) {
		return Reconciliation{}, NewFault(FaultInvalid)
	}
	result := Reconciliation{accountedCount: accounted}
	if latest != nil {
		value := latest.UTC()
		if value.IsZero() {
			return Reconciliation{}, NewFault(FaultInvalid)
		}
		result.latestHitAt = &value
	}
	return result, nil
}

func (result Reconciliation) AccountedCount() int64 { return result.accountedCount }
func (result Reconciliation) LatestHitAt() (time.Time, bool) {
	if result.latestHitAt == nil {
		return time.Time{}, false
	}
	return *result.latestHitAt, true
}
