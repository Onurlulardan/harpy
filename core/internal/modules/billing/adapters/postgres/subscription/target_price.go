package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	"github.com/google/uuid"
)

type targetPriceRecord struct {
	price   subscriptionapp.TargetPrice
	cadence billingperiod.Cadence
}

func (storage *store) findTargetPrice(
	ctx context.Context,
	offeringID, priceID uuid.UUID,
) (*targetPriceRecord, error) {
	return storage.findPrice(ctx, findTargetPriceSQL, offeringID, priceID)
}

func (storage *store) findPriceVersion(
	ctx context.Context,
	offeringID, priceID uuid.UUID,
) (*targetPriceRecord, error) {
	return storage.findPrice(ctx, findPriceVersionSQL, offeringID, priceID)
}

func (storage *store) findPrice(
	ctx context.Context,
	statement string,
	offeringID, priceID uuid.UUID,
) (*targetPriceRecord, error) {
	if offeringID == uuid.Nil && priceID == uuid.Nil {
		return nil, nil
	}
	var record targetPriceRecord
	var cadence string
	err := storage.queryRow(ctx, statement, offeringID, priceID).Scan(
		&record.price.OfferingVersionID, &record.price.PriceVersionID,
		&record.price.Currency, &record.price.TrialDays, &cadence,
	)
	if err != nil {
		if storage.notFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("find billing subscription target price: %w", err)
	}
	record.cadence = billingperiod.Cadence(cadence)
	if record.cadence != billingperiod.Monthly && record.cadence != billingperiod.Yearly {
		return nil, fmt.Errorf("decode billing subscription target cadence %q", cadence)
	}
	return &record, nil
}
