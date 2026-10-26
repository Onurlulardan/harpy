package storagecatalog

type queryFieldSpec struct {
	name, column, folded, kind string
	selectable, sortable       bool
	searchable, nullable       bool
}

type queryResourceSpec struct {
	name, view, operation string
	fields                []queryFieldSpec
	stable                string
}

func queryField(name, kind string, sortable bool) queryFieldSpec {
	return queryFieldSpec{name: name, column: name, kind: kind,
		selectable: true, sortable: sortable}
}

func queryNullable(name, kind string, sortable bool) queryFieldSpec {
	value := queryField(name, kind, sortable)
	value.nullable = true
	return value
}

func querySearch(name string) queryFieldSpec {
	return queryFieldSpec{name: name, column: name, folded: name + "_folded",
		kind: "search_text", selectable: true, searchable: true}
}

func queryVisibility(name string) queryFieldSpec {
	return queryFieldSpec{name: name, column: name, kind: "text_list"}
}

func coreQuerySpecs() []queryResourceSpec {
	return []queryResourceSpec{
		{name: "boards", view: "work_board_query_v1", operation: "listWorkBoards",
			stable: "id", fields: []queryFieldSpec{
				queryField("id", "uuid", true), querySearch("name"),
				queryField("description", "text", false),
				queryField("creator_id", "uuid", true),
				queryField("lifecycle", "text", true),
				queryField("row_version", "integer", true),
				queryField("task_count", "integer", true),
				queryField("member_count", "integer", true),
				queryField("created_at", "timestamp", true),
				queryField("updated_at", "timestamp", true),
				queryVisibility("visibility_principals"),
			}},
		{name: "tasks", view: "work_task_query_v1", operation: "listWorkTasks",
			stable: "id", fields: []queryFieldSpec{
				queryField("id", "uuid", true), queryField("board_id", "uuid", true),
				queryField("column_id", "uuid", true), querySearch("title"),
				queryField("description", "text", false),
				queryField("kind", "text", true), queryField("status", "text", true),
				queryField("priority", "text", true),
				queryField("creator_id", "uuid", true),
				queryField("position", "integer", true),
				queryField("row_version", "integer", true),
				queryNullable("starts_at", "timestamp", true),
				queryNullable("due_at", "timestamp", true),
				queryField("created_at", "timestamp", true),
				queryField("updated_at", "timestamp", true),
				queryVisibility("visibility_principals"),
			}},
		{name: "comments", view: "work_comment_query_v1", operation: "listWorkComments",
			stable: "id", fields: []queryFieldSpec{
				queryField("id", "uuid", true), queryField("board_id", "uuid", true),
				queryField("task_id", "uuid", true), queryField("author_id", "uuid", true),
				querySearch("body"), queryField("state", "text", true),
				queryNullable("redaction_kind", "text", true),
				queryField("row_version", "integer", true),
				queryField("created_at", "timestamp", true),
				queryNullable("redacted_at", "timestamp", true),
				queryVisibility("visibility_principals"),
			}},
	}
}
