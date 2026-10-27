package billingv1adoptionapp

import "time"

type adoptionCode string

func (code adoptionCode) Error() string { return string(code) }

func offeringRetiredAt(
	deletedAt *time.Time,
	createdAt time.Time,
) (*time.Time, bool) {
	if deletedAt == nil {
		return nil, true
	}
	if !canonicalInstant(*deletedAt) || deletedAt.Before(createdAt) {
		return nil, false
	}
	value := *deletedAt
	return &value, true
}

func sourceUpdatedAt(createdAt time.Time, updatedAt *time.Time) (time.Time, bool) {
	if updatedAt == nil {
		return createdAt, true
	}
	return *updatedAt, canonicalInstant(*updatedAt) && !updatedAt.Before(createdAt)
}

func offeringPrices(sourceID string, result PlanResult) []PriceRecord {
	prices := result.Version.Prices()
	resultRows := make([]PriceRecord, 0, len(prices))
	for _, price := range prices {
		cadence := string(price.Cadence())
		amount := price.Amount()
		resultRows = append(resultRows, PriceRecord{
			ID: mustIdentity("price-version", sourceID, cadence), Cadence: cadence,
			MinorUnits: amount.MinorUnits(), Currency: amount.Currency(),
		})
	}
	return resultRows
}

func offeringGrants(result PlanResult) []GrantRecord {
	grants := result.Version.Grants()
	rows := make([]GrantRecord, 0, len(grants))
	for _, grant := range grants {
		value := grant.Value()
		row := GrantRecord{EntitlementKey: string(grant.Definition().Key()),
			ValueType: string(grant.Definition().ValueKind()), Unlimited: value.Unlimited()}
		if boolean, ok := value.Boolean(); ok {
			row.BooleanValue = &boolean
		}
		if integer, ok := value.Integer(); ok {
			row.IntegerValue = &integer
		}
		rows = append(rows, row)
	}
	return rows
}

var _ error = adoptionCode("")
