package billingv1adoptionapp

import (
	"encoding/json"
	"sort"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

const maximumLegacyEventMetadataBytes = 64 * 1024

var legacyEventTypes = map[string]struct{}{
	"subscription_created": {}, "subscription_upgraded": {},
	"subscription_downgraded": {}, "subscription_cancelled": {},
	"subscription_resumed": {}, "subscription_renewed": {},
	"subscription_expired": {}, "trial_started": {}, "trial_ended": {},
	"plan_changed": {}, "coupon_applied": {}, "coupon_removed": {},
	"payment_succeeded": {}, "payment_failed": {},
}

func compileLegacyEvents(
	snapshot TenantSnapshot,
	subscriptions map[uuid.UUID]SubscriptionRecord,
	issues []Issue,
) ([]LegacyEventRecord, []Issue) {
	result := make([]LegacyEventRecord, 0, len(snapshot.Events))
	seen := make(map[uuid.UUID]struct{}, len(snapshot.Events))
	for _, source := range snapshot.Events {
		key := source.ID.String()
		if source.ID == uuid.Nil || source.TenantID != snapshot.TenantID {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(),
				"billing_events", "scope_invalid", key))
			continue
		}
		if _, found := seen[source.ID]; found {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(),
				"billing_events", "id_duplicate", key))
			continue
		}
		seen[source.ID] = struct{}{}
		record, err := convertLegacyEvent(source, subscriptions)
		if err != nil {
			issues = append(issues, issue("tenant", snapshot.TenantID.String(),
				"billing_events", err.Error(), key))
			continue
		}
		result = append(result, record)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SourceID < result[j].SourceID })
	return result, issues
}

func convertLegacyEvent(
	source SourceBillingEvent,
	subscriptions map[uuid.UUID]SubscriptionRecord,
) (LegacyEventRecord, error) {
	if _, found := legacyEventTypes[source.EventType]; !found ||
		!canonicalInstant(source.CreatedAt) || !validLegacyCurrency(source.Currency) {
		return LegacyEventRecord{}, adoptionCode("snapshot_invalid")
	}
	var metadata json.RawMessage
	var err error
	if len(source.Metadata) != 0 {
		metadata, err = canonicaljsonv1.Canonicalize(
			source.Metadata, maximumLegacyEventMetadataBytes,
		)
		if err != nil {
			return LegacyEventRecord{}, adoptionCode("metadata_invalid")
		}
	}
	var subscriptionID *string
	var sourceSubscriptionID *string
	if source.SubscriptionID != nil {
		subscription, found := subscriptions[*source.SubscriptionID]
		if !found {
			return LegacyEventRecord{}, adoptionCode("subscription_missing")
		}
		value := subscription.ID
		subscriptionID = &value
		sourceValue := source.SubscriptionID.String()
		sourceSubscriptionID = &sourceValue
	}
	record := LegacyEventRecord{
		SourceID: source.ID.String(),
		ID:       mustIdentity("legacy-billing-event", source.ID.String()),
		TenantID: source.TenantID.String(), SubscriptionID: subscriptionID,
		SourceSubscriptionID: sourceSubscriptionID,
		EventType:            source.EventType, AmountCents: source.AmountCents,
		Currency: source.Currency, Description: source.Description,
		Metadata: append(json.RawMessage(nil), metadata...), CreatedAt: source.CreatedAt,
	}
	record.SourceSHA256, err = recordDigest(record)
	if err != nil {
		return LegacyEventRecord{}, err
	}
	return record, nil
}

func validLegacyCurrency(value string) bool {
	count := 0
	for range value {
		count++
	}
	return count > 0 && count <= 3
}
