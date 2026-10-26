package executor

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func (batch consumedBatch) Validate() error {
	if batch.record.Validate() != nil ||
		batch.hasExactCount != (batch.countMode == api.CountModeExact) {
		return ErrContract
	}
	if batch.countMode == api.CountModeNone {
		if batch.exactCount.Validate() == nil {
			return ErrContract
		}
		return nil
	}
	if batch.countMode != api.CountModeExact || batch.exactCount.Validate() != nil {
		return ErrContract
	}
	if !sameStatementBinding(batch) {
		return ErrContract
	}
	return nil
}
