package billingv1adoptionapp

import "time"

type UsageOpeningRecord struct {
	SourceID            string    `json:"source_id"`
	MeterPeriodID       string    `json:"meter_period_id"`
	EntryID             *string   `json:"entry_id"`
	TenantID            string    `json:"tenant_id"`
	EntitlementKey      string    `json:"entitlement_key"`
	PeriodStart         time.Time `json:"period_start"`
	PeriodEnd           time.Time `json:"period_end"`
	Count               int64     `json:"count"`
	FlushEvidenceSHA256 string    `json:"flush_evidence_sha256"`
}

type GaugeOpeningRecord struct {
	MeterPeriodID  string    `json:"meter_period_id"`
	EntryID        *string   `json:"entry_id"`
	TenantID       string    `json:"tenant_id"`
	EntitlementKey string    `json:"entitlement_key"`
	OwnerFactID    string    `json:"owner_fact_id"`
	Value          int64     `json:"value"`
	EvidenceSHA256 string    `json:"evidence_sha256"`
	MeasuredAt     time.Time `json:"measured_at"`
}
