package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/idempotency"
	"github.com/google/uuid"
)

func (repository *repository) CompleteHTTPIdempotency(
	ctx context.Context,
	permit HTTPIdempotencyPermit,
	response idempotency.Response,
) (idempotency.Receipt, error) {
	if !permit.valid(repository) {
		return idempotency.Receipt{}, idempotency.ErrInvalidRequest
	}
	if response.Validate() != nil {
		return idempotency.Receipt{}, idempotency.ErrInvalidResponse
	}
	if !permit.state.used.CompareAndSwap(false, true) {
		return idempotency.Receipt{}, idempotency.ErrInvalidRequest
	}
	receiptID := uuid.New()
	request := permit.request
	var completedAt, retainUntil time.Time
	err := repository.database.QueryRow(
		ctx, insertHTTPIdempotencyReceiptSQL,
		receiptID, permit.identity.TenantID(), repository.authority.groupID,
		repository.authority.groupGeneration, repository.authority.domain,
		repository.authority.generation, request.RequestSHA256(),
		string(request.RetentionClass()), permit.identity.PrincipalID(),
		permit.identity.OperationID(), permit.identity.Key(),
		request.AuthorizationVersion(), response.Status(), response.HeadersJSON(),
		response.Body(), response.ResultReference(), response.SHA256(),
		request.AsyncJobID(), request.RequestedRetainUntil(),
	).Scan(&completedAt, &retainUntil)
	if err != nil {
		return idempotency.Receipt{}, fmt.Errorf("insert HTTP idempotency receipt: %w", err)
	}
	stored := idempotency.StoredReceipt{
		ReceiptID: receiptID, RequestSHA256: request.RequestSHA256(),
		AuthorizationVersion: request.AuthorizationVersion(),
		RetentionClass:       request.RetentionClass(), Status: response.Status(),
		HeadersJSON: response.HeadersJSON(), Body: response.Body(),
		ResultReference: response.ResultReference(), ResponseSHA256: response.SHA256(),
		AsyncJobID: request.AsyncJobID(), CompletedAt: completedAt, RetainUntil: retainUntil,
	}
	receipt, err := idempotency.RestoreReceipt(stored)
	if err != nil {
		return idempotency.Receipt{}, fmt.Errorf("restore inserted receipt: %w", err)
	}
	return receipt, nil
}
