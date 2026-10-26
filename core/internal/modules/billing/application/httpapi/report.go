package httpapi

import "encoding/json"

type AuditEventView struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenant_id"`
	Kind          string `json:"kind"`
	AggregateID   string `json:"aggregate_id"`
	OccurredAt    string `json:"occurred_at"`
	CorrelationID string `json:"correlation_id"`
	raw           json.RawMessage
}

func NewAuditEventQueryView(raw json.RawMessage) AuditEventView {
	return AuditEventView{raw: append(json.RawMessage(nil), raw...)}
}

func (view AuditEventView) MarshalJSON() ([]byte, error) {
	if len(view.raw) != 0 {
		return append([]byte(nil), view.raw...), nil
	}
	type plain AuditEventView
	return json.Marshal(plain(view))
}

type RevenueView struct {
	Currency        string `json:"currency"`
	RegistryVersion int64  `json:"currency_registry_version"`
	MonthlyMinor    int64  `json:"monthly_minor_units"`
	AnnualMinor     int64  `json:"annual_minor_units"`
}

type RevenueResult struct {
	Groups []json.RawMessage `json:"groups"`
	Meta   map[string]any    `json:"meta"`
}
