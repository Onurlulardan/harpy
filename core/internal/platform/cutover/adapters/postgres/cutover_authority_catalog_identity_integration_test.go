//go:build integration

package postgres

type builtAuthorityIdentity struct {
	kind      string
	schema    string
	name      string
	arguments string
}

var builtAuthorityObjectIdentities = []builtAuthorityIdentity{
	{kind: "schema", schema: "cutover_authority", name: "cutover_authority"},
	{kind: "table", schema: "cutover_authority", name: "authority_adoption_inventory"},
	{kind: "table", schema: "cutover_authority", name: "finalization_authorizations"},
	{kind: "table", schema: "cutover_authority", name: "finalization_receipts"},
	{kind: "table", schema: "cutover_authority", name: "full_rollout_completion_receipts"},
	{kind: "function", schema: "cutover_authority", name: "catalog_definition_sha256_v1",
		arguments: "text,oid"},
	{kind: "function", schema: "cutover_authority", name: "catalog_acl_sha256_v1",
		arguments: "text,oid"},
	{kind: "function", schema: "cutover_authority", name: "assert_adopted_catalog_v1",
		arguments: "text"},
	{kind: "function", schema: "cutover_authority", name: "finalization_state_sha256_v1",
		arguments: "uuid,uuid"},
	{kind: "function", schema: "cutover_authority", name: "record_credential_receipts_v1",
		arguments: "uuid,uuid,text,bytea,text,bytea"},
	{kind: "function", schema: "cutover_authority", name: "observe_finalizer_authority_v1",
		arguments: "uuid,uuid"},
	{kind: "function", schema: "cutover_authority", name: "assert_completed_evidence_v1",
		arguments: "uuid,uuid,bytea"},
	{kind: "function", schema: "cutover_authority", name: "assert_tenant_completion_v1",
		arguments: "uuid,uuid"},
	{kind: "function", schema: "cutover_authority", name: "assert_terminal_evidence_v1",
		arguments: "uuid,uuid,text,bytea"},
	{kind: "function", schema: "cutover_authority", name: "validate_finalization_envelope_v1",
		arguments: "uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{kind: "function", schema: "cutover_authority", name: "authorize_finalization_v1",
		arguments: "uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{kind: "function", schema: "cutover_authority", name: "consume_finalization_authorization_v1",
		arguments: "uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{kind: "function", schema: "cutover_authority", name: "finalize_operation_v1",
		arguments: "uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{kind: "function", schema: "cutover_authority", name: "verify_finalization_receipt_v1",
		arguments: "uuid,uuid,uuid,text,text"},
	{kind: "function", schema: "cutover_authority", name: "record_full_rollout_completion_v1",
		arguments: "uuid,uuid,uuid,text,text,text,text,text,text,text,text,text,timestampwithtimezone,timestampwithtimezone,timestampwithtimezone"},
}

var builtAuthorityDependencyIdentities = []builtAuthorityIdentity{
	{kind: "schema", schema: "platform", name: "platform"},
	{kind: "table", schema: "platform", name: "cutover_operations"},
	{kind: "table", schema: "platform", name: "cutover_operation_relations"},
	{kind: "table", schema: "platform", name: "cutover_operation_checkpoint_keys"},
	{kind: "table", schema: "platform", name: "cutover_operation_columns"},
	{kind: "table", schema: "platform", name: "cutover_operation_procedures"},
	{kind: "table", schema: "platform", name: "cutover_initial_copy_receipts"},
	{kind: "table", schema: "platform", name: "cutover_batch_receipts"},
	{kind: "table", schema: "platform", name: "cutover_copy_checkpoints"},
	{kind: "table", schema: "platform", name: "applied_change_receipts"},
	{kind: "table", schema: "platform", name: "cutover_change_journal"},
	{kind: "table", schema: "platform", name: "cutover_credential_receipts"},
	{kind: "table", schema: "platform", name: "tenant_migration_requirement_catalogs"},
	{kind: "table", schema: "platform", name: "tenant_migration_requirements"},
	{kind: "table", schema: "platform", name: "tenant_migration_requirement_targets"},
	{kind: "table", schema: "platform", name: "tenant_migration_checkpoints"},
	{kind: "table", schema: "platform", name: "tenants"},
	{kind: "table", schema: "platform", name: "release_generations"},
	{kind: "table", schema: "platform", name: "schema_contracts"},
}
