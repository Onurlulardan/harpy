package postgres

import _ "embed"

//go:embed sql/read_table_column_count.sql
var schemaExecutorReadTableColumnCountSQL string

//go:embed sql/read_table_constraint_count.sql
var schemaExecutorReadTableConstraintCountSQL string

//go:embed sql/read_table_index_count.sql
var schemaExecutorReadTableIndexCountSQL string

//go:embed sql/read_key_contract.sql
var schemaExecutorReadKeyContractSQL string

//go:embed sql/read_record_acl.sql
var schemaExecutorReadRecordACLSQL string
