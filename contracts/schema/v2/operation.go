package schemav2

type OperationID string

const (
	ActivateGeneration    OperationID = "activate_generation"
	AddConstraint         OperationID = "add_constraint"
	AddFieldColumn        OperationID = "add_field_column"
	AddRelationTable      OperationID = "add_relation_table"
	BackfillFieldBatch    OperationID = "backfill_field_batch"
	CleanupInvalidIndex   OperationID = "cleanup_invalid_index"
	CreateFormTable       OperationID = "create_form_table"
	CreateIndex           OperationID = "create_index"
	CreateIndexConcurrent OperationID = "create_index_concurrently"
	DisableOldWriter      OperationID = "disable_old_writer"
	DropConstraint        OperationID = "drop_constraint"
	DropFieldColumn       OperationID = "drop_field_column"
	DropFormTable         OperationID = "drop_form_table"
	DropIndex             OperationID = "drop_index"
	DropIndexConcurrent   OperationID = "drop_index_concurrently"
	DropRelationTable     OperationID = "drop_relation_table"
	EnableDualWrite       OperationID = "enable_dual_write"
	RetainGeneration      OperationID = "retain_generation"
	SetFieldNotNull       OperationID = "set_field_not_null"
	ValidateConstraint    OperationID = "validate_constraint"
	VerifyStorage         OperationID = "verify_storage"
)

type TransactionClass string

const (
	Transactional   TransactionClass = "transactional"
	ConcurrentIndex TransactionClass = "concurrent_index"
)

type ExecutorKind string

const (
	MigratorExecutor     ExecutorKind = "migrator"
	FormsRuntimeExecutor ExecutorKind = "forms_runtime"
)
