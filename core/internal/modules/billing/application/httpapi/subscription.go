package httpapi

import "encoding/json"

type SubscriptionView struct {
	ID                 string       `json:"id"`
	LegacyID           string       `json:"legacy_id"`
	TenantID           string       `json:"tenant_id"`
	LegacyPlanID       string       `json:"legacy_plan_id"`
	OfferingVersionID  string       `json:"offering_version_id"`
	PriceVersionID     string       `json:"price_version_id"`
	Provider           string       `json:"provider"`
	State              string       `json:"state"`
	PeriodStart        string       `json:"period_start"`
	PeriodEnd          string       `json:"period_end"`
	TrialEndsAt        *string      `json:"trial_ends_at,omitempty"`
	TrialStartedAt     *string      `json:"trial_started_at,omitempty"`
	GraceEndsAt        *string      `json:"grace_ends_at,omitempty"`
	CancelAtPeriodEnd  bool         `json:"cancel_at_period_end"`
	Version            int64        `json:"version"`
	EffectiveMinor     int64        `json:"effective_price_minor_units"`
	EffectiveCurrency  string       `json:"effective_price_currency"`
	EffectiveRegistry  int64        `json:"effective_price_registry_version"`
	CreatedAt          string       `json:"created_at"`
	UpdatedAt          string       `json:"updated_at"`
	CancelledAt        *string      `json:"cancelled_at,omitempty"`
	CancellationReason string       `json:"cancellation_reason,omitempty"`
	Offering           OfferingView `json:"offering"`
	raw                json.RawMessage
}

func NewSubscriptionQueryView(raw json.RawMessage) SubscriptionView {
	return SubscriptionView{raw: append(json.RawMessage(nil), raw...)}
}

func (view SubscriptionView) MarshalJSON() ([]byte, error) {
	if len(view.raw) != 0 {
		return append([]byte(nil), view.raw...), nil
	}
	type plain SubscriptionView
	return json.Marshal(plain(view))
}

type SubscriptionAccess struct {
	Access Access
}

type ScheduleChangeCommand struct {
	Access                  Access
	IdempotencyKey          string
	SubscriptionID          string
	TargetOfferingVersionID string
	TargetPriceVersionID    string
	ExpectedVersion         int64
}

type ScheduleCancellationCommand struct {
	Access          Access
	IdempotencyKey  string
	SubscriptionID  string
	ExpectedVersion int64
	Reason          string
}

type ResumeCancellationCommand struct {
	Access          Access
	IdempotencyKey  string
	SubscriptionID  string
	ExpectedVersion int64
}

type TransitionSubscriptionCommand struct {
	Access          Access
	IdempotencyKey  string
	SubscriptionID  string
	ExpectedVersion int64
	TargetState     string
	Reason          string
}
