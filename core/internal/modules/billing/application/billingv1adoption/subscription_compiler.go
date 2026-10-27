package billingv1adoptionapp

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

var subscriptionStates = map[string]struct{}{
	"pending": {}, "trialing": {}, "active": {}, "past_due": {},
	"paused": {}, "cancelled": {}, "expired": {},
}

func compileSubscriptions(
	snapshot TenantSnapshot,
	mapping TenantMapping,
	global GlobalIndex,
	issues []Issue,
) ([]SubscriptionRecord, map[uuid.UUID]SubscriptionRecord, []Issue) {
	result := make([]SubscriptionRecord, 0, len(snapshot.Subscriptions))
	index := make(map[uuid.UUID]SubscriptionRecord, len(snapshot.Subscriptions))
	grantCandidates := 0
	for _, source := range snapshot.Subscriptions {
		key := source.ID.String()
		if source.ID == uuid.Nil || source.TenantID != snapshot.TenantID {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "subscriptions", "scope_invalid", key))
			continue
		}
		if _, exists := index[source.ID]; exists {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "subscriptions", "id_duplicate", key))
			continue
		}
		offering, exists := global.Offerings[source.PlanID]
		if !exists {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "subscriptions", "plan_missing", key))
			continue
		}
		record, err := convertSubscription(source, mapping, global, offering)
		if err != nil {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(), "subscriptions", err.Error(), key))
			continue
		}
		if source.State == "trialing" || source.State == "active" || source.State == "past_due" {
			grantCandidates++
		}
		result = append(result, record)
		index[source.ID] = record
	}
	if grantCandidates > 1 {
		issues = append(issues, issue("tenant", snapshot.TenantID.String(),
			"subscriptions", "multiple_grant_candidates", snapshot.TenantID.String()))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SourceID < result[j].SourceID })
	return result, index, issues
}

func convertSubscription(
	source SourceSubscription,
	mapping TenantMapping,
	global GlobalIndex,
	offering OfferingIndex,
) (SubscriptionRecord, error) {
	if source.DeletedAt != nil || source.CreatedAt == nil || source.UpdatedAt == nil ||
		!canonicalInstant(*source.CreatedAt) || !canonicalInstant(*source.UpdatedAt) ||
		source.UpdatedAt.Before(*source.CreatedAt) || !canonicalInstant(source.PeriodStart) ||
		!canonicalInstant(source.PeriodEnd) || !source.PeriodStart.Before(source.PeriodEnd) ||
		!canonicalOptional(source.TrialStart) || !canonicalOptional(source.TrialEnd) ||
		!canonicalOptional(source.CancelledAt) || source.EffectivePriceCents < 0 {
		return SubscriptionRecord{}, adoptionCode("snapshot_invalid")
	}
	if _, exists := subscriptionStates[source.State]; !exists {
		return SubscriptionRecord{}, adoptionCode("state_invalid")
	}
	var grace *time.Time
	if source.State == "past_due" {
		value, exists := mapping.PastDueGraceEndsAt[source.ID]
		if !exists || !canonicalInstant(value) || !source.PeriodEnd.Before(value) {
			return SubscriptionRecord{}, adoptionCode("past_due_grace_unproven")
		}
		grace = &value
	}
	if source.State == "trialing" && (source.TrialEnd == nil ||
		!source.PeriodStart.Before(*source.TrialEnd)) {
		return SubscriptionRecord{}, adoptionCode("trial_boundary_invalid")
	}
	var cancelAt *time.Time
	if source.CancelAtPeriodEnd {
		value := source.PeriodEnd
		cancelAt = &value
	}
	return SubscriptionRecord{SourceID: source.ID.String(),
		ID: mustIdentity("subscription", source.ID.String()), TenantID: source.TenantID.String(),
		ProviderAccountID: global.ManualProviderAccountID,
		OfferingVersionID: offering.OfferingVersionID, PriceVersionID: offering.MonthlyPriceID,
		State: source.State, PeriodStart: source.PeriodStart, PeriodEnd: source.PeriodEnd,
		TrialStartedAt: source.TrialStart, TrialEndsAt: source.TrialEnd, GraceEndsAt: grace,
		CancelAtPeriodEnd: source.CancelAtPeriodEnd, CancelAt: cancelAt,
		CancelledAt: source.CancelledAt, CancellationReason: source.CancellationReason,
		EffectiveMinorUnits: source.EffectivePriceCents,
		CreatedAt:           *source.CreatedAt, UpdatedAt: *source.UpdatedAt,
	}, nil
}
