package attestation

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

func testBootstrapOperation(roles, privileges, ownership string) postgresv1.BootstrapOperation {
	return postgresv1.BootstrapOperation{
		ContractVersion: "postgres-bootstrap-operation.v1",
		DatabaseOrigin:  "fresh", DatabaseName: "harpy",
		InputHashes: postgresv1.BootstrapInputHashes{
			RolesManifestSHA256: roles, PrivilegesManifestSHA256: privileges,
			OwnershipManifestSHA256: ownership,
		},
		Steps: []postgresv1.OperationStep{
			testStep("verify_artifact_hashes", "read_only_preflight"),
			testStep("verify_credential_lease", "read_only_preflight"),
			testStep("verify_external_cutover_authority_owner", "read_only_preflight"),
			testStep("create_database", "autocommit_ddl"),
			testStep("reconcile_roles", "transactional_ddl"),
			testStep("reconcile_memberships", "transactional_ddl"),
			testStep("reconcile_privileges", "transactional_ddl"),
			testStep("verify_ownership_inventory", "read_only_preflight"),
			testStep("verify_final_catalog", "read_only_preflight"),
			testStep("publish_bootstrap_receipt", "receipt_publication"),
		},
	}
}

func testExtensionOperation(extension, ownership string) postgresv1.ExtensionOperation {
	return postgresv1.ExtensionOperation{
		ContractVersion: "postgres-extension-operation.v1",
		DatabaseOrigin:  "fresh", DatabaseName: "harpy",
		InputHashes: postgresv1.ExtensionInputHashes{
			ExtensionManifestSHA256: extension, OwnershipManifestSHA256: ownership,
		},
		Steps: []postgresv1.OperationStep{
			testStep("verify_artifact_hashes", "read_only_preflight"),
			testStep("verify_credential_lease", "read_only_preflight"),
			testStep("verify_bootstrap_receipt", "read_only_preflight"),
			testStep("resolve_control_files", "read_only_preflight"),
			testStep("reconcile_extensions", "transactional_ddl"),
			testStep("verify_extension_catalog", "read_only_preflight"),
			testStep("verify_ownership_inventory", "read_only_preflight"),
			testStep("publish_extension_receipt", "receipt_publication"),
		},
	}
}

func testCleanupOperation(roles, privileges string) postgresv1.CredentialCleanupOperation {
	return postgresv1.CredentialCleanupOperation{
		ContractVersion: "credential-cleanup-operation.v1",
		DatabaseOrigin:  "fresh", DatabaseName: "harpy",
		InputHashes: postgresv1.CredentialCleanupHashes{
			RolesManifestSHA256: roles, PrivilegesManifestSHA256: privileges,
		},
		PrivilegedRoles: []string{"harpy_cluster_admin", "harpy_extension_admin"},
		Steps: []postgresv1.OperationStep{
			testStep("verify_artifact_hashes", "read_only_preflight"),
			testStep("disable_privileged_login", "external_provisioner"),
			testStep("clear_privileged_password", "external_provisioner"),
			testStep("terminate_privileged_sessions", "external_provisioner"),
			testStep("remove_privileged_memberships", "external_provisioner"),
			testStep("verify_final_role_catalog", "external_provisioner"),
			testStep("revoke_mounted_credentials", "external_provisioner"),
			testStep("publish_cleanup_receipt", "receipt_publication"),
		},
	}
}

func testStep(id, class string) postgresv1.OperationStep {
	return postgresv1.OperationStep{StepID: id, TransactionClass: class}
}
