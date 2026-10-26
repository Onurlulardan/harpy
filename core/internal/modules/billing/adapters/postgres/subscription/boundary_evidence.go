package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	"github.com/google/uuid"
)

type boundaryEvidence struct {
	outcome     string
	paymentID   dbUUID
	periodStart dbInstant
	periodEnd   dbInstant
	graceEndsAt dbInstant
}

func (storage *store) findBoundaryEvidence(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.JobExecution,
	job assertedJob,
) (*boundaryEvidence, error) {
	var evidence boundaryEvidence
	err := storage.queryRow(
		ctx, findBoundaryEvidenceSQL, authorityID, storage.tenantID,
		execution.JobID, execution.SubscriptionID, string(execution.Boundary),
		job.runAt, execution.ExpectedRevision,
	).Scan(
		&evidence.outcome, &evidence.paymentID, &evidence.periodStart,
		&evidence.periodEnd, &evidence.graceEndsAt,
	)
	if err != nil {
		if storage.notFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("find billing boundary evidence: %w", err)
	}
	return &evidence, nil
}

func (evidence *boundaryEvidence) apply(
	boundary *subscriptionapp.BoundarySnapshot,
) error {
	if evidence == nil || evidence.outcome == "unpaid" {
		return nil
	}
	if evidence.outcome == "grace" {
		boundary.GraceEndsAt = instantPointer(evidence.graceEndsAt)
		return nil
	}
	if evidence.outcome != "paid" || !evidence.paymentID.Valid ||
		!evidence.periodStart.Valid || !evidence.periodEnd.Valid {
		return fmt.Errorf("invalid billing boundary evidence outcome")
	}
	period, err := billingperiod.New(
		evidence.periodStart.Time.UTC(), evidence.periodEnd.Time.UTC(),
	)
	if err != nil {
		return fmt.Errorf("decode billing boundary evidence period: %w", err)
	}
	paymentID := evidence.paymentID.UUID
	boundary.PaymentReceiptID, boundary.NextPeriod = &paymentID, &period
	return nil
}
