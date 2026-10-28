package main

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

func bootstrapOperation(roles, privileges, ownership string) postgresv1.BootstrapOperation {
	return postgresv1.BootstrapOperation{
		ContractVersion: "postgres-bootstrap-operation.v1", DatabaseOrigin: "fresh",
		DatabaseName: databaseName, InputHashes: postgresv1.BootstrapInputHashes{
			RolesManifestSHA256: roles, PrivilegesManifestSHA256: privileges,
			OwnershipManifestSHA256: ownership,
		},
		Steps: operationSteps(
			[]string{
				"verify_artifact_hashes", "verify_credential_lease",
				"verify_external_cutover_authority_owner", "create_database", "reconcile_roles",
				"reconcile_memberships", "reconcile_privileges", "verify_ownership_inventory",
				"verify_final_catalog", "publish_bootstrap_receipt",
			},
			[]string{
				"read_only_preflight", "read_only_preflight", "read_only_preflight",
				"autocommit_ddl", "transactional_ddl", "transactional_ddl",
				"transactional_ddl", "read_only_preflight", "read_only_preflight",
				"receipt_publication",
			},
		),
	}
}

func extensionOperation(extension, ownership string) postgresv1.ExtensionOperation {
	return postgresv1.ExtensionOperation{
		ContractVersion: "postgres-extension-operation.v1", DatabaseOrigin: "fresh",
		DatabaseName: databaseName, InputHashes: postgresv1.ExtensionInputHashes{
			ExtensionManifestSHA256: extension, OwnershipManifestSHA256: ownership,
		},
		Steps: operationSteps(
			[]string{
				"verify_artifact_hashes", "verify_credential_lease", "verify_bootstrap_receipt",
				"resolve_control_files", "reconcile_extensions", "verify_extension_catalog",
				"verify_ownership_inventory", "publish_extension_receipt",
			},
			[]string{
				"read_only_preflight", "read_only_preflight", "read_only_preflight",
				"read_only_preflight", "transactional_ddl", "read_only_preflight",
				"read_only_preflight", "receipt_publication",
			},
		),
	}
}

func cleanupOperation(roles, privileges string) postgresv1.CredentialCleanupOperation {
	return postgresv1.CredentialCleanupOperation{
		ContractVersion: "credential-cleanup-operation.v1", DatabaseOrigin: "fresh",
		DatabaseName: databaseName, InputHashes: postgresv1.CredentialCleanupHashes{
			RolesManifestSHA256: roles, PrivilegesManifestSHA256: privileges,
		},
		PrivilegedRoles: []string{"harpy_cluster_admin", "harpy_extension_admin"},
		Steps: operationSteps(
			[]string{
				"verify_artifact_hashes", "disable_privileged_login", "clear_privileged_password",
				"terminate_privileged_sessions", "remove_privileged_memberships",
				"verify_final_role_catalog", "revoke_mounted_credentials",
				"publish_cleanup_receipt",
			},
			[]string{
				"read_only_preflight", "external_provisioner", "external_provisioner",
				"external_provisioner", "external_provisioner", "external_provisioner",
				"external_provisioner", "receipt_publication",
			},
		),
	}
}

func operationSteps(names, classes []string) []postgresv1.OperationStep {
	steps := make([]postgresv1.OperationStep, len(names))
	for index := range names {
		steps[index] = postgresv1.OperationStep{
			StepID: names[index], TransactionClass: classes[index],
		}
	}
	return steps
}
