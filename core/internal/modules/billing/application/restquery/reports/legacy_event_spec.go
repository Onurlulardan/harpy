package reports

func legacyEventSpec() resourceSpec {
	fields := []fieldSpec{
		{"b2520000-0000-4000-8000-000000000001", "id", "id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2520000-0000-4000-8000-000000000002", "tenant_id", "tenant_id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2520000-0000-4000-8000-000000000003", "subscription_id", "subscription_id", "uuid", nullableScalarFilters, true, 36, true, false, nil},
		{"b2520000-0000-4000-8000-000000000004", "event_type", "event_type", "text", scalarFilters, true, 127, true, false, nil},
		{"b2520000-0000-4000-8000-000000000005", "amount_cents", "amount_cents", "integer", orderedFilters, true, 20, true, false, nil},
		{"b2520000-0000-4000-8000-000000000006", "currency", "currency", "text", scalarFilters, true, 3, true, false, nil},
		{"b2520000-0000-4000-8000-000000000007", "description", "description", "text", scalarFilters, false, 0, true, false, nil},
		{"b2520000-0000-4000-8000-000000000008", "metadata", "metadata", "json", nil, false, 0, true, false, nil},
		{"b2520000-0000-4000-8000-000000000009", "created_at", "created_at", "timestamp", orderedFilters, true, 27, true, false, nil},
	}
	return resourceSpec{
		Name:       LegacyEventResource,
		ResourceID: "b2500000-0000-4000-8000-000000000001",
		VersionID:  "b2510000-0000-4000-8000-000000000001",
		View:       "billing_v1_event_report_query_v1", StableID: fields[0].ID,
		Operation: AuditOperation, Fields: fields,
	}
}
