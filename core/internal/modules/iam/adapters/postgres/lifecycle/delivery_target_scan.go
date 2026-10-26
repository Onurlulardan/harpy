package postgres

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func scanDeliveryTarget(row lifecycleRow) (lifecycle.DeliveryTarget, error) {
	var result lifecycle.DeliveryTarget
	err := row.Scan(&result.SubjectUserID, &result.ResourceID,
		&result.RecipientIdentifierID, &result.RecipientAddress)
	if err != nil {
		return lifecycle.DeliveryTarget{}, mapError(err)
	}
	if result.Validate() != nil {
		return lifecycle.DeliveryTarget{}, lifecycle.ErrUnavailable
	}
	return result, nil
}
