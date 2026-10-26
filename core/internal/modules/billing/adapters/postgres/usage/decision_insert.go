package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

func (storage *store) writeDecision(
	ctx context.Context,
	decision usageapp.DecisionRecord,
) error {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return err
	}
	var limitAmount any
	if value, ok := decision.Limit.Integer(); ok {
		limitAmount = value
	}
	var insertedID uuid.UUID
	err = storage.queryRow(ctx, insertDecisionSQL, authorityID, storage.tenantID,
		storage.actorID, decision.ID, string(decision.EntitlementKey),
		decision.PeriodID, decision.SourceOperationID, decision.Requested,
		decision.Current, limitAmount, decision.Limit.Unlimited(), decision.Allowed,
		decision.SoftExceeded, decision.DecidedAt,
	).Scan(&insertedID)
	if err != nil {
		return fmt.Errorf("insert billing entitlement decision: %w", err)
	}
	if insertedID != decision.ID {
		return usageapp.ErrSnapshot
	}
	return nil
}
