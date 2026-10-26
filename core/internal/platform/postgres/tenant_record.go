package postgres

type TenantRecord struct {
	schemaName string
	slug       string
}

func (record TenantRecord) SchemaName() string { return record.schemaName }
func (record TenantRecord) Slug() string       { return record.slug }
