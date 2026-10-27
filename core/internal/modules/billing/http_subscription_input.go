package billing

type scheduleChangeInput struct {
	TargetOfferingVersionID string `json:"target_offering_version_id"`
	TargetPriceVersionID    string `json:"target_price_version_id"`
	ExpectedVersion         int64  `json:"expected_version"`
}

type cancellationInput struct {
	ExpectedVersion int64  `json:"expected_version"`
	Reason          string `json:"reason"`
}

type resumeCancellationInput struct {
	ExpectedVersion int64 `json:"expected_version"`
}

type discountRedemptionInput struct {
	Code string `json:"code"`
}

type discountValidationInput struct {
	Code              string `json:"code"`
	OfferingVersionID string `json:"offering_version_id"`
}

type transitionInput struct {
	ExpectedVersion int64  `json:"expected_version"`
	TargetState     string `json:"target_state"`
	Reason          string `json:"reason"`
}
