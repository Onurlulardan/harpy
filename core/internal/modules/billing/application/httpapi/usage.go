package httpapi

type UsageMeterView struct {
	EntitlementKey string  `json:"entitlement_key"`
	Kind           string  `json:"kind"`
	Current        int64   `json:"current"`
	Limit          *int64  `json:"limit"`
	Unlimited      bool    `json:"unlimited"`
	PeriodStart    *string `json:"period_start,omitempty"`
	PeriodEnd      *string `json:"period_end,omitempty"`
	ResetPolicy    string  `json:"reset_policy"`
}

type UsageView struct {
	Meters []UsageMeterView `json:"meters"`
}
