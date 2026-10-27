package presentation

func defaultWidgetDescriptors() []WidgetDescriptor {
	all := []string{
		"boolean", "date", "decimal", "document_ref", "document_ref_list", "enum",
		"enum_list", "integer", "json", "relation", "text", "text_list", "time",
		"timestamp", "uuid", "uuid_list",
	}
	return []WidgetDescriptor{
		{ID: "auto", FieldTypes: all},
		{ID: "checkbox", FieldTypes: []string{"boolean"}},
		{ID: "date", FieldTypes: []string{"date"}},
		{ID: "datetime", FieldTypes: []string{"timestamp"}},
		{ID: "document_selector", FieldTypes: []string{"document_ref", "document_ref_list"}},
		{ID: "document_uploader", FieldTypes: []string{"document_ref", "document_ref_list"}},
		{ID: "email", FieldTypes: []string{"text"}},
		{ID: "json", FieldTypes: []string{"json"}},
		{ID: "lookup", FieldTypes: []string{"relation", "uuid", "uuid_list"}},
		{ID: "multi_select", FieldTypes: []string{"enum_list", "text_list", "uuid_list"}},
		{ID: "number", FieldTypes: []string{"decimal", "integer"}},
		{ID: "password", FieldTypes: []string{"text"}},
		{ID: "radio", FieldTypes: []string{"enum"}},
		{ID: "select", FieldTypes: []string{"enum"}},
		{ID: "telephone", FieldTypes: []string{"text"}},
		{ID: "text", FieldTypes: []string{"text"}},
		{ID: "textarea", FieldTypes: []string{"text"}},
		{ID: "time", FieldTypes: []string{"time"}},
		{ID: "toggle", FieldTypes: []string{"boolean"}},
		{ID: "url", FieldTypes: []string{"text"}},
	}
}
