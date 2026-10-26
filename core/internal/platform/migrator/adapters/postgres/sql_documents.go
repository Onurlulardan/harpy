package postgres

import _ "embed"

var (
	//go:embed sql/table_exists.sql
	tableExistsSQL string
	//go:embed sql/create_schema.sql
	createSchemaSQL string
	//go:embed sql/create_history_table.sql
	createHistoryTableSQL string
	//go:embed sql/create_reject_mutation_function.sql
	createRejectMutationFunctionSQL string
	//go:embed sql/create_reject_mutation_trigger.sql
	createRejectMutationTriggerSQL string
	//go:embed sql/insert_anchor.sql
	insertAnchorSQL string
	//go:embed sql/insert_migration.sql
	insertMigrationSQL string
	//go:embed sql/get_migration.sql
	getMigrationSQL string
	//go:embed sql/get_latest_version.sql
	getLatestVersionSQL string
	//go:embed sql/list_migrations.sql
	listMigrationsSQL string
)
