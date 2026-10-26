package subscription

import (
	"context"
	"fmt"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
	"github.com/google/uuid"
)

type snapshotRow struct {
	id, tenantID, providerID uuid.UUID
	offeringID, priceID      uuid.UUID
	currency, state          string
	periodStart, periodEnd   dbInstant
	trialEnd, graceEnd       dbInstant
	cancelAtEnd              bool
	cancelAt                 dbInstant
	revision                 int64
	updatedAt                time.Time
}

func scanSnapshot(scanner row) (subscriptionapp.SubscriptionSnapshot, error) {
	var record snapshotRow
	err := scanner.Scan(
		&record.id, &record.tenantID, &record.providerID,
		&record.offeringID, &record.priceID, &record.currency, &record.state,
		&record.periodStart, &record.periodEnd, &record.trialEnd, &record.graceEnd,
		&record.cancelAtEnd, &record.cancelAt, &record.revision, &record.updatedAt,
	)
	if err != nil {
		return subscriptionapp.SubscriptionSnapshot{}, err
	}
	return record.snapshot()
}

func (record snapshotRow) snapshot() (subscriptionapp.SubscriptionSnapshot, error) {
	period, err := record.period()
	if err != nil {
		return subscriptionapp.SubscriptionSnapshot{}, err
	}
	return subscriptionapp.SubscriptionSnapshot{
		ID: record.id, TenantID: record.tenantID,
		ProviderAccountID: record.providerID,
		OfferingVersionID: record.offeringID, PriceVersionID: record.priceID,
		Currency: record.currency, State: subscriptiondomain.State(record.state),
		Period: period, TrialEndsAt: instantPointer(record.trialEnd),
		GraceEndsAt:       instantPointer(record.graceEnd),
		CancelAtPeriodEnd: record.cancelAtEnd,
		CancelAt:          instantPointer(record.cancelAt), Revision: record.revision,
		UpdatedAt: record.updatedAt.UTC(),
	}, nil
}

func (record snapshotRow) period() (*billingperiod.Period, error) {
	if !record.periodStart.Valid && !record.periodEnd.Valid {
		return nil, nil
	}
	if !record.periodStart.Valid || !record.periodEnd.Valid {
		return nil, fmt.Errorf("partial billing subscription period")
	}
	value, err := billingperiod.New(
		record.periodStart.Time.UTC(), record.periodEnd.Time.UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("decode billing subscription period: %w", err)
	}
	return &value, nil
}

func (storage *store) lockSubscription(
	ctx context.Context,
	authorityID, subscriptionID uuid.UUID,
) (subscriptionapp.SubscriptionSnapshot, error) {
	result, err := scanSnapshot(storage.queryRow(
		ctx, lockSubscriptionSQL, authorityID, storage.tenantID, subscriptionID,
	))
	if err != nil {
		return subscriptionapp.SubscriptionSnapshot{}, fmt.Errorf(
			"lock billing subscription: %w", err,
		)
	}
	return result, nil
}
