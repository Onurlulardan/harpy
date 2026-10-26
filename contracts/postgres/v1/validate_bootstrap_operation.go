package postgresv1

import "fmt"

func (operation BootstrapOperation) Validate() error {
	if !operationHeaderValid(
		operation.ContractVersion, "postgres-bootstrap-operation.v1",
		operation.DatabaseOrigin, operation.DatabaseName,
	) {
		return fmt.Errorf("bootstrap operation header is invalid")
	}
	if err := validateInputDigests(
		operation.InputHashes.RolesManifestSHA256,
		operation.InputHashes.PrivilegesManifestSHA256,
		operation.InputHashes.OwnershipManifestSHA256,
	); err != nil {
		return err
	}
	databaseStep := step("create_database", "autocommit_ddl")
	if DatabaseOrigin(operation.DatabaseOrigin) == DatabaseOriginV1Adopted {
		databaseStep = step("verify_adopted_database", "read_only_preflight")
	}
	expected := []OperationStep{
		step("verify_artifact_hashes", "read_only_preflight"),
		step("verify_credential_lease", "read_only_preflight"),
		step("verify_external_cutover_authority_owner", "read_only_preflight"),
		databaseStep,
		step("reconcile_roles", "transactional_ddl"),
		step("reconcile_memberships", "transactional_ddl"),
		step("reconcile_privileges", "transactional_ddl"),
		step("verify_ownership_inventory", "read_only_preflight"),
		step("verify_final_catalog", "read_only_preflight"),
		step("publish_bootstrap_receipt", "receipt_publication"),
	}
	return validateOperationSteps(operation.Steps, expected)
}
