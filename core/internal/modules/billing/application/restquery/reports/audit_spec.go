package reports

func auditSpec() resourceSpec {
	fields := []fieldSpec{
		{"b2320000-0000-4000-8000-000000000001", "id", "id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2320000-0000-4000-8000-000000000002", "tenant_id", "tenant_id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2320000-0000-4000-8000-000000000003", "kind", "kind", "text", scalarFilters, true, 127, true, false, nil},
		{"b2320000-0000-4000-8000-000000000004", "aggregate_id", "aggregate_id", "uuid", scalarFilters, true, 36, true, false, nil},
		{"b2320000-0000-4000-8000-000000000005", "occurred_at", "occurred_at", "timestamp", orderedFilters, true, 27, true, false, nil},
		{"b2320000-0000-4000-8000-000000000006", "correlation_id", "correlation_id", "uuid", scalarFilters, true, 36, true, false, nil},
	}
	return resourceSpec{
		Name: AuditResource, ResourceID: "b2300000-0000-4000-8000-000000000001",
		VersionID: "b2310000-0000-4000-8000-000000000001",
		View:      "billing_audit_report_query_v1", StableID: fields[0].ID,
		Operation: AuditOperation, Fields: fields,
	}
}
