package postgresv1

import "fmt"

func (operation ExtensionOperation) Validate() error {
	if !operationHeaderValid(
		operation.ContractVersion, "postgres-extension-operation.v1",
		operation.DatabaseOrigin, operation.DatabaseName,
	) {
		return fmt.Errorf("extension operation header is invalid")
	}
	if err := validateInputDigests(
		operation.InputHashes.ExtensionManifestSHA256,
		operation.InputHashes.OwnershipManifestSHA256,
	); err != nil {
		return err
	}
	expected := []OperationStep{
		step("verify_artifact_hashes", "read_only_preflight"),
		step("verify_credential_lease", "read_only_preflight"),
		step("verify_bootstrap_receipt", "read_only_preflight"),
		step("resolve_control_files", "read_only_preflight"),
		step("reconcile_extensions", "transactional_ddl"),
		step("verify_extension_catalog", "read_only_preflight"),
		step("verify_ownership_inventory", "read_only_preflight"),
		step("publish_extension_receipt", "receipt_publication"),
	}
	return validateOperationSteps(operation.Steps, expected)
}
