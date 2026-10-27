package billingv1adoptionapp

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

func compileUsageOpenings(
	snapshot TenantSnapshot,
	mapping TenantMapping,
	issues []Issue,
) ([]UsageOpeningRecord, []Issue) {
	result := make([]UsageOpeningRecord, 0, len(snapshot.Usage))
	seen := map[string]struct{}{}
	for _, source := range snapshot.Usage {
		key := source.ID.String()
		_, _, monthValid := monthPeriod(source.Month)
		evidence := mapping.UsageFlushEvidenceSHA256[source.ID]
		if source.ID == uuid.Nil || source.TenantID != snapshot.TenantID || source.Count < 0 ||
			!monthValid || !canonicalOptional(source.CreatedAt) || !canonicalOptional(source.UpdatedAt) ||
			!validDigest(evidence) {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "submission_usage", "opening_evidence_invalid", key))
			continue
		}
		if _, exists := seen[source.Month]; exists {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "submission_usage", "month_duplicate", source.Month))
			continue
		}
		seen[source.Month] = struct{}{}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].PeriodStart.Before(result[j].PeriodStart) })
	return result, issues
}

var requiredGaugeKeys = []string{
	"documents.storage_bytes", "iam.active_users",
}

func compileGaugeOpenings(
	snapshot TenantSnapshot,
	issues []Issue,
) ([]GaugeOpeningRecord, []Issue) {
	result := make([]GaugeOpeningRecord, 0, len(snapshot.Gauges))
	seen := map[string]struct{}{}
	for _, source := range snapshot.Gauges {
		if source.Value < 0 || source.OwnerFactID == uuid.Nil || !validDigest(source.EvidenceSHA256) ||
			!canonicalInstant(source.MeasuredAt) || source.MeasuredAt.After(snapshot.SnapshotAt) {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "gauge_openings", "owner_evidence_invalid", source.EntitlementKey))
			continue
		}
		if _, exists := seen[source.EntitlementKey]; exists {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "gauge_openings", "owner_evidence_duplicate", source.EntitlementKey))
			continue
		}
		seen[source.EntitlementKey] = struct{}{}
		row := GaugeOpeningRecord{MeterPeriodID: mustIdentity("gauge-period",
			snapshot.TenantID.String(), source.EntitlementKey), TenantID: snapshot.TenantID.String(),
			EntitlementKey: source.EntitlementKey, OwnerFactID: source.OwnerFactID.String(),
			Value: source.Value, EvidenceSHA256: source.EvidenceSHA256, MeasuredAt: source.MeasuredAt}
		if source.Value > 0 {
			value := mustIdentity("gauge-opening", source.OwnerFactID.String())
			row.EntryID = &value
		}
		result = append(result, row)
	}
	for _, key := range requiredGaugeKeys {
		if _, exists := seen[key]; !exists {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "gauge_openings", "owner_evidence_missing", key))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].EntitlementKey < result[j].EntitlementKey })
	return result, issues
}

var gaugePeriodEnd = time.Date(9999, 12, 31, 23, 59, 59, 999999000, time.UTC)
