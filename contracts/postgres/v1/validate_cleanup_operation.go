package postgresv1

import "fmt"

func (operation CredentialCleanupOperation) Validate() error {
	if !operationHeaderValid(
		operation.ContractVersion, "credential-cleanup-operation.v1",
		operation.DatabaseOrigin, operation.DatabaseName,
	) {
		return fmt.Errorf("credential cleanup operation header is invalid")
	}
	if err := validateInputDigests(
		operation.InputHashes.RolesManifestSHA256,
		operation.InputHashes.PrivilegesManifestSHA256,
	); err != nil {
		return err
	}
	if len(operation.PrivilegedRoles) != 2 ||
		operation.PrivilegedRoles[0] != "harpy_cluster_admin" ||
		operation.PrivilegedRoles[1] != "harpy_extension_admin" {
		return fmt.Errorf("privileged role order is invalid")
	}
	expected := []OperationStep{
		step("verify_artifact_hashes", "read_only_preflight"),
		step("disable_privileged_login", "external_provisioner"),
		step("clear_privileged_password", "external_provisioner"),
		step("terminate_privileged_sessions", "external_provisioner"),
		step("remove_privileged_memberships", "external_provisioner"),
		step("verify_final_role_catalog", "external_provisioner"),
		step("revoke_mounted_credentials", "external_provisioner"),
		step("publish_cleanup_receipt", "receipt_publication"),
	}
	return validateOperationSteps(operation.Steps, expected)
}
