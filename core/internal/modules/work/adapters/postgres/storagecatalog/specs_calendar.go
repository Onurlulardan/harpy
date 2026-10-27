package storagecatalog

func calendarQuerySpecs() []queryResourceSpec {
	return []queryResourceSpec{
		{name: "calendar", view: "work_calendar_event_query_v1",
			operation: "listWorkCalendarEvents", stable: "id",
			fields: calendarEventFields()},
		{name: "inbox", view: "work_inbox_query_v1", operation: "listWorkInbox",
			stable: "id", fields: projectionQueryFields(false)},
		{name: "calendar-unified", view: "work_calendar_query_v1",
			operation: "listWorkCalendar", stable: "id",
			fields: projectionQueryFields(true)},
	}
}

func calendarEventFields() []queryFieldSpec {
	return []queryFieldSpec{
		queryField("id", "uuid", true), queryField("owner_id", "uuid", true),
		queryField("audience", "text", true), queryNullable("board_id", "uuid", true),
		queryNullable("organization_id", "uuid", true), querySearch("title"),
		queryField("description", "text", false), queryField("color", "text", true),
		queryNullable("starts_at", "timestamp", true),
		queryNullable("ends_at", "timestamp", true),
		queryNullable("start_date", "date", true), queryNullable("end_date", "date", true),
		queryField("timezone", "text", true), queryField("generated", "boolean", true),
		queryField("row_version", "integer", true),
		queryField("created_at", "timestamp", true),
		queryField("updated_at", "timestamp", true),
		queryVisibility("visibility_principals"),
	}
}

func projectionQueryFields(calendar bool) []queryFieldSpec {
	fields := []queryFieldSpec{
		queryField("id", "uuid", true), queryField("source_owner", "text", true),
		queryField("source_kind", "text", true), queryField("source_id", "uuid", true),
		queryField("source_version_id", "uuid", true),
		queryField("source_revision", "integer", true),
		queryField("permission_revision", "integer", true), querySearch("summary"),
		queryNullable("due_at", "timestamp", true),
		queryNullable("board_id", "uuid", true), queryNullable("status", "text", true),
		queryNullable("priority", "text", true),
		queryField("updated_at", "timestamp", true),
		queryVisibility("visibility_principals"),
		queryVisibility("context_principals"),
	}
	if calendar {
		fields = append(fields,
			queryField("item_kind", "text", true),
			queryNullable("starts_at", "timestamp", true),
			queryNullable("ends_at", "timestamp", true),
			queryNullable("start_date", "date", true),
			queryNullable("end_date", "date", true))
	}
	return fields
}

func allQuerySpecs() []queryResourceSpec {
	return append(coreQuerySpecs(), calendarQuerySpecs()...)
}
