package httpapi

import "encoding/json"

type LegacyEventView struct {
	ID             string           `json:"id"`
	TenantID       string           `json:"tenant_id"`
	SubscriptionID *string          `json:"subscription_id"`
	EventType      string           `json:"event_type"`
	AmountCents    int64            `json:"amount_cents"`
	Currency       string           `json:"currency"`
	Description    string           `json:"description"`
	Metadata       *json.RawMessage `json:"metadata"`
	CreatedAt      string           `json:"created_at"`
	raw            json.RawMessage
}

func NewLegacyEventQueryView(raw json.RawMessage) LegacyEventView {
	return LegacyEventView{raw: append(json.RawMessage(nil), raw...)}
}

func (view LegacyEventView) MarshalJSON() ([]byte, error) {
	if len(view.raw) != 0 {
		return append([]byte(nil), view.raw...), nil
	}
	type plain LegacyEventView
	return json.Marshal(plain(view))
}
