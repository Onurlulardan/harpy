package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/idempotency"
	"github.com/google/uuid"
)

func (repository *repository) readHTTPIdempotency(
	ctx context.Context,
	identity idempotency.Identity,
) (bool, *idempotency.Receipt, error) {
	var locked bool
	var receiptID, requestSHA, retention sql.NullString
	var authorizationVersion, status sql.NullInt64
	var headers, body []byte
	var resultReference, responseSHA, asyncJobID sql.NullString
	var completedAt, retainUntil sql.NullTime
	err := repository.database.QueryRow(
		ctx, readHTTPIdempotencyReceiptSQL, identity.AdvisoryKey(),
		identity.TenantID(), identity.PrincipalID(), identity.OperationID(), identity.Key(),
	).Scan(
		&locked, &receiptID, &requestSHA, &authorizationVersion,
		&retention, &status, &headers, &body,
		&resultReference, &responseSHA, &asyncJobID, &completedAt, &retainUntil,
	)
	if err != nil {
		return false, nil, err
	}
	if !locked || !receiptID.Valid {
		return locked, nil, nil
	}
	parsedReceiptID, err := uuid.Parse(receiptID.String)
	if err != nil {
		return false, nil, fmt.Errorf("parse idempotency receipt ID: %w", err)
	}
	stored := idempotency.StoredReceipt{
		ReceiptID: parsedReceiptID, RequestSHA256: requestSHA.String,
		AuthorizationVersion: authorizationVersion.Int64,
		RetentionClass:       idempotency.RetentionClass(retention.String),
		Status:               int(status.Int64), HeadersJSON: headers, Body: body,
		ResponseSHA256: responseSHA.String,
		CompletedAt:    completedAt.Time, RetainUntil: retainUntil.Time,
	}
	if resultReference.Valid {
		stored.ResultReference = &resultReference.String
	}
	if asyncJobID.Valid {
		value, parseErr := uuid.Parse(asyncJobID.String)
		if parseErr != nil {
			return false, nil, fmt.Errorf("parse idempotency job ID: %w", parseErr)
		}
		stored.AsyncJobID = &value
	}
	receipt, err := idempotency.RestoreReceipt(stored)
	if err != nil {
		return false, nil, fmt.Errorf("restore HTTP idempotency receipt: %w", err)
	}
	return true, &receipt, nil
}
