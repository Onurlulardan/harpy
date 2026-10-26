package subscriptionapp

import "github.com/google/uuid"

func checkedResult(result Result, subscriptionID uuid.UUID) (Result, error) {
	if !result.Subscription.valid() || result.Subscription.ID != subscriptionID {
		return Result{}, ErrSnapshot
	}
	return result, nil
}
