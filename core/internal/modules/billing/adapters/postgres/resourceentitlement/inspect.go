package resourceentitlement

import (
	"context"
	"fmt"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	"github.com/google/uuid"
)

func (storage *store) Inspect(
	ctx context.Context,
	request billingapi.MeteredResourceMutationRequest,
) (resourceapp.Inspection, error) {
	if !storage.usable(ctx) || !request.Valid() ||
		request.TenantID != storage.tenantID || request.ActorID != storage.actorID {
		return resourceapp.Inspection{}, resourceapp.ErrRequest
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return resourceapp.Inspection{}, err
	}
	var receiptFound, featureEnabled, allowed, unlimited bool
	var operationID, ownerFactID *uuid.UUID
	var requestHash []byte
	var gaugeDelta, current, limit int64
	var recordedAt *time.Time
	var projectionFound bool
	var projectionState string
	var projectionRevision int64
	var projectionUpdatedAt *time.Time
	err = storage.queryRow(ctx, inspectSQL, authorityID, storage.tenantID,
		storage.actorID, string(request.Resource), string(request.Mutation),
		request.IdempotencyKey, request.ProposedOwnerFactID,
	).Scan(&receiptFound, &operationID, &ownerFactID, &requestHash,
		&featureEnabled, &allowed, &gaugeDelta, &current, &limit, &unlimited,
		&recordedAt, &projectionFound, &projectionState, &projectionRevision,
		&projectionUpdatedAt)
	if err != nil {
		return resourceapp.Inspection{}, fmt.Errorf("inspect metered resource: %w", err)
	}
	return decodeInspection(request, receiptFound, operationID, ownerFactID,
		requestHash, featureEnabled, allowed, gaugeDelta, current, limit, unlimited,
		recordedAt, projectionFound, projectionState, projectionRevision,
		projectionUpdatedAt)
}

func decodeInspection(
	request billingapi.MeteredResourceMutationRequest,
	receiptFound bool, operationID, ownerFactID *uuid.UUID, requestHash []byte,
	featureEnabled, allowed bool, gaugeDelta, current, limit int64, unlimited bool,
	recordedAt *time.Time, projectionFound bool, projectionState string,
	projectionRevision int64, projectionUpdatedAt *time.Time,
) (resourceapp.Inspection, error) {
	result := resourceapp.Inspection{}
	if receiptFound {
		if operationID == nil || ownerFactID == nil || len(requestHash) != 32 ||
			recordedAt == nil {
			return result, resourceapp.ErrUnavailable
		}
		var digest [32]byte
		copy(digest[:], requestHash)
		result.Stored = &resourceapp.StoredMutation{Decision: billingapi.MeteredResourceMutationDecision{
			Resource: request.Resource, Mutation: request.Mutation,
			TenantID: request.TenantID, ActorID: request.ActorID,
			OperationID: *operationID, OwnerFactID: *ownerFactID,
			IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
			FeatureEnabled: featureEnabled, Allowed: allowed,
			GaugeDelta: gaugeDelta, Current: current, Limit: limit,
			Unlimited: unlimited,
		}, RecordedAt: *recordedAt}
	}
	if projectionFound {
		if projectionUpdatedAt == nil {
			return resourceapp.Inspection{}, resourceapp.ErrUnavailable
		}
		result.Projection = &resourceapp.ResourceProjection{
			Resource: request.Resource, OwnerFactID: request.ProposedOwnerFactID,
			State:    resourceapp.CountedState(projectionState),
			Revision: projectionRevision, UpdatedAt: *projectionUpdatedAt,
		}
	}
	return result, nil
}
