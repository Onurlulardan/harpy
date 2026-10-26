package reports

func subscriptionSpec() resourceSpec {
	fields := []fieldSpec{
		{"b2220000-0000-4000-8000-000000000001", "id", "id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2220000-0000-4000-8000-000000000002", "tenant_id", "tenant_id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2220000-0000-4000-8000-000000000003", "offering_version_id", "offering_version_id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2220000-0000-4000-8000-000000000004", "price_version_id", "price_version_id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2220000-0000-4000-8000-000000000005", "provider", "provider", "text", scalarFilters, true, 63, true, false, nil},
		{"b2220000-0000-4000-8000-000000000006", "state", "state", "text", scalarFilters, true, 16, true, false, nil},
		{"b2220000-0000-4000-8000-000000000007", "period_start", "period_start", "timestamp", nullableOrderedFilters, true, 27, true, false, nil},
		{"b2220000-0000-4000-8000-000000000008", "period_end", "period_end", "timestamp", nullableOrderedFilters, true, 27, true, false, nil},
		{"b2220000-0000-4000-8000-000000000009", "trial_ends_at", "trial_ends_at", "timestamp", nullableOrderedFilters, true, 27, true, false, nil},
		{"b2220000-0000-4000-8000-00000000000a", "grace_ends_at", "grace_ends_at", "timestamp", nullableOrderedFilters, true, 27, true, false, nil},
		{"b2220000-0000-4000-8000-00000000000b", "cancel_at_period_end", "cancel_at_period_end", "boolean", scalarFilters, true, 5, true, false, nil},
		{"b2220000-0000-4000-8000-00000000000c", "version", "version", "integer", orderedFilters, true, 20, true, false, nil},
		{"b2220000-0000-4000-8000-00000000000d", "effective_price_minor_units", "effective_price_minor_units", "integer", orderedFilters, true, 20, true, false, nil},
		{"b2220000-0000-4000-8000-00000000000e", "effective_price_currency", "effective_price_currency", "text", scalarFilters, true, 3, true, false, nil},
		{"b2220000-0000-4000-8000-00000000000f", "effective_price_registry_version", "effective_price_registry_version", "integer", orderedFilters, true, 20, true, false, nil},
		{"b2220000-0000-4000-8000-000000000010", "created_at", "created_at", "timestamp", orderedFilters, true, 27, true, false, nil},
		{"b2220000-0000-4000-8000-000000000011", "updated_at", "updated_at", "timestamp", orderedFilters, true, 27, true, false, nil},
		{"b2220000-0000-4000-8000-000000000012", "offering", "offering", "json", nil, false, 0, true, false, nil},
		{"b2220000-0000-4000-8000-000000000013", "legacy_id", "legacy_id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2220000-0000-4000-8000-000000000014", "legacy_plan_id", "legacy_plan_id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2220000-0000-4000-8000-000000000015", "trial_started_at", "trial_started_at", "timestamp", nullableOrderedFilters, true, 27, true, false, nil},
		{"b2220000-0000-4000-8000-000000000016", "cancelled_at", "cancelled_at", "timestamp", nullableOrderedFilters, true, 27, true, false, nil},
		{"b2220000-0000-4000-8000-000000000017", "cancellation_reason", "cancellation_reason", "text", nullableScalarFilters, false, 0, true, false, nil},
	}
	return resourceSpec{
		Name: SubscriptionResource, ResourceID: "b2200000-0000-4000-8000-000000000001",
		VersionID: "b2210000-0000-4000-8000-000000000001",
		View:      "billing_subscription_report_query_v2", StableID: fields[0].ID,
		Operation: SubscriptionOperation, Fields: fields,
	}
}
