package resourceentitlement

import (
	"context"
	"fmt"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	"github.com/google/uuid"
)

func (storage *store) RecordReconciliation(
	ctx context.Context,
	record resourceapp.ReconciliationRecord,
) error {
	decision := record.Decision
	request := reconciliationRequest(decision)
	query := billingapi.MeteredResourceEvidenceQuery{
		Resource: decision.Resource, TenantID: decision.TenantID,
		OwnerFactID: decision.OwnerFactID,
	}
	if !storage.usable(ctx) || decision.Replay || !request.Valid() ||
		!record.Evidence.ValidFor(query) || !decision.ValidFor(request, record.Evidence) ||
		decision.TenantID != storage.tenantID || decision.ActorID != storage.actorID ||
		!validRecordTime(record.RecordedAt) || !validReconciliationLink(record) {
		return resourceapp.ErrRequest
	}
	beforeRevision, err := validateReconciliationChange(record)
	if err != nil {
		return err
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return err
	}
	var recorded bool
	err = storage.queryRow(ctx, recordReconciliationSQL, authorityID, storage.tenantID,
		storage.actorID, string(decision.Resource), decision.OperationID,
		decision.OwnerFactID, decision.IdempotencyKey, decision.RequestSHA256[:],
		decision.EvidenceSHA256[:], decision.Exists, decision.Active,
		decision.IncludedInOpening, decision.Changed, decision.GaugeDelta,
		decision.Current, decision.Limit, decision.Unlimited,
		nullUUID(record.UsageDecisionID), nullUUID(record.UsageEntryID),
		beforeRevision, record.RecordedAt, record.Evidence.MeasuredAt,
	).Scan(&recorded)
	if err != nil {
		return fmt.Errorf("record metered resource reconciliation: %w", err)
	}
	if !recorded {
		return resourceapp.ErrUnavailable
	}
	return nil
}

func reconciliationRequest(
	decision billingapi.MeteredResourceReconcileDecision,
) billingapi.MeteredResourceReconcileRequest {
	return billingapi.MeteredResourceReconcileRequest{
		Resource: decision.Resource, TenantID: decision.TenantID,
		ActorID: decision.ActorID, ProposedOperationID: decision.OperationID,
		OwnerFactID: decision.OwnerFactID, IdempotencyKey: decision.IdempotencyKey,
		RequestSHA256: decision.RequestSHA256,
	}
}

func validReconciliationLink(record resourceapp.ReconciliationRecord) bool {
	if record.Decision.GaugeDelta != 0 {
		return validV7(record.UsageDecisionID) && validV7(record.UsageEntryID)
	}
	return record.UsageDecisionID == uuid.Nil && record.UsageEntryID == uuid.Nil
}
