package cutoverv1

var authorityDependencyOrder = [][4]string{
	{"schema", "platform", "platform", ""},
	{"table", "platform", "cutover_operations", ""},
	{"table", "platform", "cutover_operation_relations", ""},
	{"table", "platform", "cutover_operation_checkpoint_keys", ""},
	{"table", "platform", "cutover_operation_columns", ""},
	{"table", "platform", "cutover_operation_procedures", ""},
	{"table", "platform", "cutover_initial_copy_receipts", ""},
	{"table", "platform", "cutover_batch_receipts", ""},
	{"table", "platform", "cutover_copy_checkpoints", ""},
	{"table", "platform", "applied_change_receipts", ""},
	{"table", "platform", "cutover_change_journal", ""},
	{"table", "platform", "cutover_credential_receipts", ""},
	{"table", "platform", "tenant_migration_requirement_catalogs", ""},
	{"table", "platform", "tenant_migration_requirements", ""},
	{"table", "platform", "tenant_migration_requirement_targets", ""},
	{"table", "platform", "tenant_migration_checkpoints", ""},
	{"table", "platform", "tenants", ""},
	{"table", "platform", "release_generations", ""},
	{"table", "platform", "schema_contracts", ""},
}

func validateAuthorityDependencies(
	dependencies []AuthorityCatalogDependency, contractSHA256 string,
) error {
	if len(dependencies) != len(authorityDependencyOrder) ||
		contractSHA256 != AuthorityDependencyContractSHA256(dependencies) {
		return contractError("authority dependency catalog")
	}
	for index, dependency := range dependencies {
		expected := authorityDependencyOrder[index]
		if dependency.Kind != expected[0] || dependency.SchemaName != expected[1] ||
			dependency.ObjectName != expected[2] || dependency.IdentityArguments != expected[3] ||
			dependency.Owner != "harpy_owner" ||
			!validDigest(dependency.DefinitionSHA256) || !validDigest(dependency.ACLSHA256) {
			return contractError("authority dependency")
		}
	}
	return nil
}

func validateAuthorityReceiptDependencies(value AuthorityAdoptionReceipt) error {
	if len(value.Dependencies) != len(authorityDependencyOrder) ||
		value.DependencyContractSHA256 !=
			AuthorityReceiptDependencyContractSHA256(value.Dependencies) ||
		value.DependencyInstanceSHA256 !=
			AuthorityReceiptDependencyInstanceSHA256(value.Dependencies) {
		return contractError("authority receipt dependency catalog")
	}
	dependencies := make([]AuthorityCatalogDependency, len(value.Dependencies))
	for index, dependency := range value.Dependencies {
		if dependency.ObjectOID == 0 {
			return contractError("authority receipt dependency OID")
		}
		dependencies[index] = dependency.AuthorityCatalogDependency
	}
	return validateAuthorityDependencies(dependencies, value.DependencyContractSHA256)
}
