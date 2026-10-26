package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/idempotency"
	"github.com/google/uuid"
)

// BeginAuthorizedHTTPIdempotency must be called only after the current
// actor/resource/response-field authorization has been refreshed in this scope.
func (repository *repository) BeginAuthorizedHTTPIdempotency(
	ctx context.Context,
	request idempotency.Request,
) (HTTPIdempotencyDecision, error) {
	if repository == nil || request.Validate() != nil || repository.principalID == uuid.Nil {
		return HTTPIdempotencyDecision{}, idempotency.ErrInvalidRequest
	}
	identity, err := idempotency.NewIdentity(
		repository.authority.scopeID, repository.principalID,
		repository.operationID, request.Key(),
	)
	if err != nil {
		return HTTPIdempotencyDecision{}, err
	}
	locked, receipt, err := repository.readHTTPIdempotency(ctx, identity)
	if err != nil {
		return HTTPIdempotencyDecision{}, fmt.Errorf("inspect HTTP idempotency identity: %w", err)
	}
	if !locked {
		return HTTPIdempotencyDecision{}, idempotency.ErrInProgress
	}
	if receipt == nil {
		return permitDecision(HTTPIdempotencyPermit{
			writer: repository, identity: identity, request: request,
			state: &httpIdempotencyPermitState{},
		}), nil
	}
	if receipt.RequestSHA256() != request.RequestSHA256() ||
		receipt.RetentionClass() != request.RetentionClass() {
		return HTTPIdempotencyDecision{}, idempotency.ErrKeyReused
	}
	return replayDecision(*receipt), nil
}
