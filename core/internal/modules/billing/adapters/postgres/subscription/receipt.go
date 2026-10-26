package subscription

import (
	"context"
	"fmt"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type receiptRecord struct {
	receipt        subscriptionapp.OperationReceipt
	subscriptionID uuid.UUID
	sourceKind     string
}

func (record *receiptRecord) matches(
	execution subscriptionapp.Execution,
	expectedSource string,
) bool {
	return record != nil && record.sourceKind == expectedSource &&
		record.subscriptionID == execution.SubscriptionID &&
		record.receipt.SourceOperationID == execution.SourceOperationID
}

func (storage *store) findReceipt(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.Execution,
) (*receiptRecord, error) {
	var id, subscriptionID, sourceID uuid.UUID
	var sourceKind, operationID, currency, state string
	var requestBytes, resultBytes []byte
	var providerID, offeringID, priceID uuid.UUID
	var periodStart, periodEnd, trialEnd, graceEnd dbInstant
	var cancelAt dbInstant
	var cancelAtEnd bool
	var revision int64
	var updatedAt, completedAt time.Time
	err := storage.queryRow(
		ctx, findReceiptSQL, authorityID, storage.tenantID, execution.SourceOperationID,
	).Scan(
		&id, &subscriptionID, &sourceID, &sourceKind, &operationID,
		&requestBytes, &resultBytes, &providerID, &offeringID, &priceID,
		&currency, &state, &periodStart, &periodEnd, &trialEnd, &graceEnd,
		&cancelAtEnd, &cancelAt, &revision, &updatedAt, &completedAt,
	)
	if err != nil {
		if storage.notFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("find billing subscription receipt: %w", err)
	}
	requestDigest, err := requiredDigest(requestBytes)
	if err != nil {
		return nil, fmt.Errorf("decode billing subscription request digest: %w", err)
	}
	resultDigest, err := requiredDigest(resultBytes)
	if err != nil {
		return nil, fmt.Errorf("decode billing subscription result digest: %w", err)
	}
	snapshot, err := receiptSnapshot(receiptSnapshotInput{
		id: subscriptionID, tenantID: storage.tenantID, providerID: providerID,
		offeringID: offeringID, priceID: priceID, currency: currency, state: state,
		periodStart: periodStart, periodEnd: periodEnd, trialEnd: trialEnd,
		graceEnd: graceEnd, cancelAtEnd: cancelAtEnd, cancelAt: cancelAt,
		revision: revision, updatedAt: updatedAt,
	})
	if err != nil {
		return nil, err
	}
	return &receiptRecord{
		subscriptionID: subscriptionID, sourceKind: sourceKind,
		receipt: subscriptionapp.OperationReceipt{
			ID: id, OperationID: operationID, SourceOperationID: sourceID,
			RequestSHA256: requestDigest, ResultSHA256: resultDigest,
			Result: snapshot, CompletedAt: completedAt.UTC(),
		},
	}, nil
}

type receiptSnapshotInput struct {
	id, tenantID, providerID, offeringID, priceID uuid.UUID
	currency, state                               string
	periodStart, periodEnd                        dbInstant
	trialEnd, graceEnd, cancelAt                  dbInstant
	cancelAtEnd                                   bool
	revision                                      int64
	updatedAt                                     time.Time
}

func receiptSnapshot(input receiptSnapshotInput) (subscriptionapp.SubscriptionSnapshot, error) {
	record := snapshotRow{
		id: input.id, tenantID: input.tenantID, providerID: input.providerID,
		offeringID: input.offeringID, priceID: input.priceID,
		currency: input.currency, state: input.state,
		periodStart: input.periodStart, periodEnd: input.periodEnd,
		trialEnd: input.trialEnd, graceEnd: input.graceEnd,
		cancelAtEnd: input.cancelAtEnd, cancelAt: input.cancelAt,
		revision: input.revision, updatedAt: input.updatedAt,
	}
	return record.snapshot()
}
