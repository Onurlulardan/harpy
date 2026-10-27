package reports

func revenueSpec() resourceSpec {
	fields := []fieldSpec{
		{"b2420000-0000-4000-8000-000000000001", "id", "id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2420000-0000-4000-8000-000000000002", "state", "state", "text", scalarFilters, true, 16, false, false, nil},
		{"b2420000-0000-4000-8000-000000000003", "cadence", "cadence", "text", scalarFilters, true, 16, false, false, nil},
		{"b2420000-0000-4000-8000-000000000004", "currency", "currency", "text", scalarFilters, true, 3, false, true, nil},
		{"b2420000-0000-4000-8000-000000000005", "currency_registry_version", "currency_registry_version", "integer", orderedFilters, true, 20, false, true, nil},
		{"b2420000-0000-4000-8000-000000000006", "subscription_revision", "subscription_revision", "integer", orderedFilters, true, 20, false, false, nil},
		{"b2420000-0000-4000-8000-000000000007", "offering_version_id", "offering_version_id", "uuid", scalarFilters, true, 36, false, false, nil},
		{"b2420000-0000-4000-8000-000000000008", "price_version_id", "price_version_id", "uuid", scalarFilters, true, 36, false, false, nil},
		{"b2420000-0000-4000-8000-000000000009", "monthly_minor_units", "monthly_minor_units", "integer", orderedFilters, true, 20, false, false, []string{"sum"}},
		{"b2420000-0000-4000-8000-00000000000a", "annual_minor_units", "annual_minor_units", "integer", orderedFilters, true, 20, false, false, []string{"sum"}},
	}
	return resourceSpec{
		Name: RevenueResource, ResourceID: "b2400000-0000-4000-8000-000000000001",
		VersionID: "b2410000-0000-4000-8000-000000000001",
		View:      "billing_revenue_report_query_v1", StableID: fields[0].ID,
		Operation: RevenueOperation, Aggregate: true, Fields: fields,
	}
}
