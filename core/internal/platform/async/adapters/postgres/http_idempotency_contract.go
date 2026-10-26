package postgres

import (
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/idempotency"
)

type httpIdempotencyPermitState struct {
	used atomic.Bool
}

type HTTPIdempotencyPermit struct {
	writer   *repository
	identity idempotency.Identity
	request  idempotency.Request
	state    *httpIdempotencyPermitState
}

func (permit HTTPIdempotencyPermit) valid(writer *repository) bool {
	return writer != nil && permit.writer == writer && permit.state != nil &&
		permit.request.Validate() == nil
}

type HTTPIdempotencyDecision struct {
	replay *idempotency.Receipt
	permit HTTPIdempotencyPermit
}

func replayDecision(receipt idempotency.Receipt) HTTPIdempotencyDecision {
	return HTTPIdempotencyDecision{replay: &receipt}
}

func permitDecision(permit HTTPIdempotencyPermit) HTTPIdempotencyDecision {
	return HTTPIdempotencyDecision{permit: permit}
}

func (decision HTTPIdempotencyDecision) Replay() (idempotency.Receipt, bool) {
	if decision.replay == nil {
		return idempotency.Receipt{}, false
	}
	return *decision.replay, true
}

func (decision HTTPIdempotencyDecision) Permit() (HTTPIdempotencyPermit, bool) {
	if decision.replay != nil || decision.permit.state == nil {
		return HTTPIdempotencyPermit{}, false
	}
	return decision.permit, true
}
