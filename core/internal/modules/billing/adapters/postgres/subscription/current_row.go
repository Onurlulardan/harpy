package subscription

import (
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
	"github.com/google/uuid"
)

type currentRow struct {
	candidateCount           int64
	id, tenantID, providerID uuid.UUID
	legacyID, legacyPlanID   uuid.UUID
	offeringID, priceID      uuid.UUID
	provider                 dbText
	state                    string
	periodStart, periodEnd   dbInstant
	trialEnd, graceEnd       dbInstant
	trialStart, cancelledAt  dbInstant
	cancellationReason       dbText
	cancelAtEnd              bool
	cancelAt                 dbInstant
	effectiveMinor           int64
	effectiveCurrency        dbText
	effectiveRegistry        int64
	revision                 int64
	createdAt, updatedAt     time.Time
	priceMatches             bool
	offeringMatches          bool
	offeringDocument         []byte
}

func scanCurrent(scanner row) (currentRow, error) {
	var record currentRow
	err := scanner.Scan(
		&record.candidateCount, &record.id, &record.legacyID, &record.tenantID,
		&record.legacyPlanID, &record.providerID,
		&record.offeringID, &record.priceID, &record.provider, &record.state,
		&record.periodStart, &record.periodEnd, &record.trialEnd, &record.graceEnd,
		&record.cancelAtEnd, &record.cancelAt,
		&record.trialStart, &record.cancelledAt, &record.cancellationReason,
		&record.effectiveMinor,
		&record.effectiveCurrency, &record.effectiveRegistry,
		&record.revision, &record.createdAt, &record.updatedAt,
		&record.priceMatches,
		&record.offeringMatches, &record.offeringDocument,
	)
	return record, err
}

func (record currentRow) value() (subscriptionapp.CurrentSubscription, error) {
	if record.candidateCount != 1 || !record.provider.Valid ||
		!record.effectiveCurrency.Valid || !record.priceMatches ||
		!record.offeringMatches {
		return subscriptionapp.CurrentSubscription{}, subscriptionapp.ErrCurrentSubscriptionInvariant
	}
	period, err := record.period()
	if err != nil {
		return subscriptionapp.CurrentSubscription{}, subscriptionapp.ErrCurrentSubscriptionInvariant
	}
	offering, err := decodeCurrentOffering(record.offeringDocument)
	if err != nil {
		return subscriptionapp.CurrentSubscription{}, subscriptionapp.ErrCurrentSubscriptionInvariant
	}
	return subscriptionapp.NewCurrentSubscription(subscriptionapp.CurrentSubscriptionInput{
		ID: record.id, LegacyID: record.legacyID, LegacyPlanID: record.legacyPlanID,
		TenantID: record.tenantID, ProviderAccountID: record.providerID,
		OfferingVersionID: record.offeringID, PriceVersionID: record.priceID,
		Provider: record.provider.Text, State: subscriptiondomain.State(record.state),
		Period: period, TrialEndsAt: instantPointer(record.trialEnd),
		TrialStartedAt:     instantPointer(record.trialStart),
		CancelledAt:        instantPointer(record.cancelledAt),
		CancellationReason: record.cancellationReason.Text,
		GraceEndsAt:        instantPointer(record.graceEnd), CancelAtPeriodEnd: record.cancelAtEnd,
		CancelAt: instantPointer(record.cancelAt), Revision: record.revision,
		EffectiveMinor:    record.effectiveMinor,
		EffectiveCurrency: record.effectiveCurrency.Text,
		EffectiveRegistry: record.effectiveRegistry,
		CreatedAt:         record.createdAt.UTC(), UpdatedAt: record.updatedAt.UTC(), Offering: offering,
	})
}

func (record currentRow) period() (*billingperiod.Period, error) {
	if !record.periodStart.Valid && !record.periodEnd.Valid {
		return nil, nil
	}
	if !record.periodStart.Valid || !record.periodEnd.Valid {
		return nil, subscriptionapp.ErrCurrentSubscriptionInvariant
	}
	period, err := billingperiod.New(record.periodStart.Time, record.periodEnd.Time)
	if err != nil {
		return nil, err
	}
	return &period, nil
}
