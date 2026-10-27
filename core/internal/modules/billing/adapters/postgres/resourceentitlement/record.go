package resourceentitlement

import (
	"context"
	"fmt"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	"github.com/google/uuid"
)

func (storage *store) Record(
	ctx context.Context,
	record resourceapp.MutationRecord,
) error {
	decision := record.Decision
	request := billingapi.MeteredResourceMutationRequest{
		Resource: decision.Resource, Mutation: decision.Mutation,
		TenantID: decision.TenantID, ActorID: decision.ActorID,
		ProposedOperationID: decision.OperationID,
		ProposedOwnerFactID: decision.OwnerFactID,
		IdempotencyKey:      decision.IdempotencyKey, RequestSHA256: decision.RequestSHA256,
	}
	if !storage.usable(ctx) || decision.Replay || !decision.ValidFor(request) ||
		decision.TenantID != storage.tenantID || decision.ActorID != storage.actorID ||
		!validRecordTime(record.RecordedAt) || !validUsageLink(record) {
		return resourceapp.ErrRequest
	}
	beforeRevision, err := validateChange(decision, record.Change)
	if err != nil {
		return err
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return err
	}
	var recorded bool
	err = storage.queryRow(ctx, recordSQL, authorityID, storage.tenantID,
		storage.actorID, string(decision.Resource), string(decision.Mutation),
		decision.OperationID, decision.OwnerFactID, decision.IdempotencyKey,
		decision.RequestSHA256[:], decision.FeatureEnabled, decision.Allowed,
		decision.GaugeDelta, decision.Current, decision.Limit, decision.Unlimited,
		nullUUID(record.UsageDecisionID), nullUUID(record.UsageEntryID),
		beforeRevision, record.RecordedAt,
	).Scan(&recorded)
	if err != nil {
		return fmt.Errorf("record metered resource: %w", err)
	}
	if !recorded {
		return resourceapp.ErrUnavailable
	}
	return nil
}

func validRecordTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1_000 == 0
}

func validUsageLink(record resourceapp.MutationRecord) bool {
	decisionID, entryID := record.UsageDecisionID, record.UsageEntryID
	if record.Decision.GaugeDelta != 0 {
		return validV7(decisionID) && validV7(entryID)
	}
	if record.Decision.Mutation == billingapi.MeteredResourceCreate ||
		record.Decision.Mutation == billingapi.MeteredResourceRestore {
		if record.Decision.FeatureEnabled && !record.Decision.Allowed {
			return validV7(decisionID) && entryID == uuid.Nil
		}
	}
	return decisionID == uuid.Nil && entryID == uuid.Nil
}

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func nullUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}
