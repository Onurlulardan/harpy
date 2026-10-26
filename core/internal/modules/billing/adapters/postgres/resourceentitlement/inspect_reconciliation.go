package resourceentitlement

import (
	"context"
	"fmt"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	"github.com/google/uuid"
)

func (storage *store) InspectReconciliation(
	ctx context.Context,
	request billingapi.MeteredResourceReconcileRequest,
) (resourceapp.ReconciliationInspection, error) {
	if !storage.usable(ctx) || !request.Valid() || request.TenantID != storage.tenantID ||
		request.ActorID != storage.actorID {
		return resourceapp.ReconciliationInspection{}, resourceapp.ErrRequest
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return resourceapp.ReconciliationInspection{}, err
	}
	var receiptFound, exists, active, included, changed, unlimited bool
	var operationID *uuid.UUID
	var requestHash, evidenceHash []byte
	var gaugeDelta, current, limit int64
	var recordedAt *time.Time
	var projectionFound bool
	var state string
	var revision int64
	var updatedAt *time.Time
	err = storage.queryRow(ctx, inspectReconciliationSQL, authorityID, storage.tenantID,
		storage.actorID, string(request.Resource), request.IdempotencyKey,
		request.OwnerFactID,
	).Scan(&receiptFound, &operationID, &requestHash, &evidenceHash,
		&exists, &active, &included, &changed, &gaugeDelta, &current, &limit,
		&unlimited, &recordedAt, &projectionFound, &state, &revision, &updatedAt)
	if err != nil {
		return resourceapp.ReconciliationInspection{},
			fmt.Errorf("inspect metered resource reconciliation: %w", err)
	}
	return decodeReconciliation(request, receiptFound, operationID, requestHash,
		evidenceHash, exists, active, included, changed, gaugeDelta, current, limit,
		unlimited, recordedAt, projectionFound, state, revision, updatedAt)
}

func decodeReconciliation(
	request billingapi.MeteredResourceReconcileRequest,
	receiptFound bool,
	operationID *uuid.UUID,
	requestHash, evidenceHash []byte,
	exists, active, included, changed bool,
	gaugeDelta, current, limit int64,
	unlimited bool,
	recordedAt *time.Time,
	projectionFound bool,
	state string,
	revision int64,
	updatedAt *time.Time,
) (resourceapp.ReconciliationInspection, error) {
	result := resourceapp.ReconciliationInspection{}
	if receiptFound {
		if operationID == nil || len(requestHash) != 32 || len(evidenceHash) != 32 ||
			recordedAt == nil {
			return result, resourceapp.ErrUnavailable
		}
		var requestDigest, evidenceDigest [32]byte
		copy(requestDigest[:], requestHash)
		copy(evidenceDigest[:], evidenceHash)
		result.Stored = &resourceapp.StoredReconciliation{Decision: billingapi.MeteredResourceReconcileDecision{
			Resource: request.Resource, TenantID: request.TenantID,
			ActorID: request.ActorID, OperationID: *operationID,
			OwnerFactID: request.OwnerFactID, IdempotencyKey: request.IdempotencyKey,
			RequestSHA256: requestDigest, EvidenceSHA256: evidenceDigest,
			Exists: exists, Active: active, IncludedInOpening: included,
			Changed: changed, GaugeDelta: gaugeDelta, Current: current,
			Limit: limit, Unlimited: unlimited,
		}, RecordedAt: *recordedAt}
	}
	if projectionFound {
		if updatedAt == nil {
			return resourceapp.ReconciliationInspection{}, resourceapp.ErrUnavailable
		}
		result.Projection = &resourceapp.ResourceProjection{
			Resource: request.Resource, OwnerFactID: request.OwnerFactID,
			State: resourceapp.CountedState(state), Revision: revision,
			UpdatedAt: *updatedAt,
		}
	}
	return result, nil
}
