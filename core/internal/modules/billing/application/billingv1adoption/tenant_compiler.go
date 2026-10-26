package billingv1adoptionapp

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

func compileTenant(
	snapshot TenantSnapshot,
	mapping TenantMapping,
	index GlobalIndex,
	timezone string,
) (ScopePlan, error) {
	if snapshot.TenantID == uuid.Nil || !canonicalInstant(snapshot.SnapshotAt) {
		return ScopePlan{}, fmt.Errorf("Billing V1 tenant snapshot identity is invalid")
	}
	plan := emptyScopePlan("tenant", snapshot.TenantID.String(), timezone, snapshot.SnapshotAt)
	subscriptions := append([]SourceSubscription(nil), snapshot.Subscriptions...)
	redemptions := append([]SourceRedemption(nil), snapshot.Redemptions...)
	usage := append([]SourceUsage(nil), snapshot.Usage...)
	gauges := append([]GaugeOpening(nil), snapshot.Gauges...)
	events := append([]SourceBillingEvent(nil), snapshot.Events...)
	sort.Slice(subscriptions, func(i, j int) bool { return subscriptions[i].ID.String() < subscriptions[j].ID.String() })
	sort.Slice(redemptions, func(i, j int) bool { return redemptions[i].ID.String() < redemptions[j].ID.String() })
	sort.Slice(usage, func(i, j int) bool { return usage[i].Month < usage[j].Month })
	sort.Slice(gauges, func(i, j int) bool { return gauges[i].EntitlementKey < gauges[j].EntitlementKey })
	sort.Slice(events, func(i, j int) bool { return events[i].ID.String() < events[j].ID.String() })
	for _, category := range []struct {
		name  string
		rows  any
		count int
	}{
		{"subscriptions", subscriptions, len(subscriptions)},
		{"subscription_coupons", redemptions, len(redemptions)},
		{"submission_usage", usage, len(usage)},
		{"gauge_openings", gauges, len(gauges)},
		{"billing_events", events, len(events)},
	} {
		evidence, err := categoryEvidence(category.name, category.rows, category.count)
		if err != nil {
			return ScopePlan{}, err
		}
		plan.Categories = append(plan.Categories, evidence)
	}
	if mapping.TenantID != snapshot.TenantID {
		plan.Issues = append(plan.Issues, issue("tenant", snapshot.TenantID.String(),
			"mapping", "tenant_mapping_missing", snapshot.TenantID.String()))
	}
	plan.BaseAssignments = compileBaseAssignment(snapshot, index, &plan.Issues)
	baseEvidence, evidenceErr := categoryEvidence(
		"base_assignments", plan.BaseAssignments, len(plan.BaseAssignments),
	)
	if evidenceErr != nil {
		return ScopePlan{}, evidenceErr
	}
	plan.Categories = append(plan.Categories, baseEvidence)
	var subscriptionIndex map[uuid.UUID]SubscriptionRecord
	plan.Subscriptions, subscriptionIndex, plan.Issues = compileSubscriptions(
		snapshot, mapping, index, plan.Issues,
	)
	plan.Redemptions, plan.Issues = compileRedemptions(
		snapshot, index, subscriptionIndex, plan.Issues,
	)
	plan.LegacyEvents, plan.Issues = compileLegacyEvents(
		snapshot, subscriptionIndex, plan.Issues,
	)
	plan.UsageOpenings, plan.Issues = compileUsageOpenings(snapshot, mapping, plan.Issues)
	plan.GaugeOpenings, plan.Issues = compileGaugeOpenings(snapshot, plan.Issues)
	validateTenantMoney(&plan, index)
	var err error
	plan.SourceSHA256, err = sourceRoot(plan.Categories)
	if err != nil {
		return ScopePlan{}, err
	}
	finalizePlan(&plan)
	return plan, nil
}

func compileBaseAssignment(
	snapshot TenantSnapshot,
	index GlobalIndex,
	issues *[]Issue,
) []BaseOfferingRecord {
	if index.Base.SourceID == uuid.Nil {
		*issues = append(*issues, issue("tenant", snapshot.TenantID.String(),
			"base_assignment", "base_plan_missing", snapshot.TenantID.String()))
		return []BaseOfferingRecord{}
	}
	tenantID := snapshot.TenantID.String()
	return []BaseOfferingRecord{{
		ID:       mustIdentity("base-assignment", tenantID, index.Base.SourceID.String()),
		TenantID: tenantID, OfferingVersionID: index.Base.OfferingVersionID,
		SourceOperationID: mustIdentity("base-assignment-source", tenantID,
			index.Base.SourceID.String()), EffectiveStart: snapshot.SnapshotAt,
	}}
}
