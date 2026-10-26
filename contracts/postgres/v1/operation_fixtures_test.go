package postgresv1

func validBootstrapOperation(origin string) map[string]any {
	databaseStep := operationStep("create_database", "autocommit_ddl")
	if origin == "v1_adopted" {
		databaseStep = operationStep("verify_adopted_database", "read_only_preflight")
	}
	return map[string]any{
		"contract_version": "postgres-bootstrap-operation.v1",
		"database_origin":  origin, "database_name": "harpy",
		"input_hashes": map[string]any{
			"roles_manifest_sha256":      digest("a"),
			"privileges_manifest_sha256": digest("b"),
			"ownership_manifest_sha256":  digest("c"),
		},
		"steps": []any{
			operationStep("verify_artifact_hashes", "read_only_preflight"),
			operationStep("verify_credential_lease", "read_only_preflight"),
			operationStep("verify_external_cutover_authority_owner", "read_only_preflight"),
			databaseStep,
			operationStep("reconcile_roles", "transactional_ddl"),
			operationStep("reconcile_memberships", "transactional_ddl"),
			operationStep("reconcile_privileges", "transactional_ddl"),
			operationStep("verify_ownership_inventory", "read_only_preflight"),
			operationStep("verify_final_catalog", "read_only_preflight"),
			operationStep("publish_bootstrap_receipt", "receipt_publication"),
		},
	}
}

func validExtensionOperation() map[string]any {
	return map[string]any{
		"contract_version": "postgres-extension-operation.v1",
		"database_origin":  "fresh", "database_name": "harpy",
		"input_hashes": map[string]any{
			"extension_manifest_sha256": digest("d"),
			"ownership_manifest_sha256": digest("e"),
		},
		"steps": []any{
			operationStep("verify_artifact_hashes", "read_only_preflight"),
			operationStep("verify_credential_lease", "read_only_preflight"),
			operationStep("verify_bootstrap_receipt", "read_only_preflight"),
			operationStep("resolve_control_files", "read_only_preflight"),
			operationStep("reconcile_extensions", "transactional_ddl"),
			operationStep("verify_extension_catalog", "read_only_preflight"),
			operationStep("verify_ownership_inventory", "read_only_preflight"),
			operationStep("publish_extension_receipt", "receipt_publication"),
		},
	}
}

func validCleanupOperation() map[string]any {
	return map[string]any{
		"contract_version": "credential-cleanup-operation.v1",
		"database_origin":  "fresh", "database_name": "harpy",
		"input_hashes": map[string]any{
			"roles_manifest_sha256":      digest("f"),
			"privileges_manifest_sha256": digest("0"),
		},
		"privileged_roles": []any{
			"harpy_cluster_admin", "harpy_extension_admin",
		},
		"steps": []any{
			operationStep("verify_artifact_hashes", "read_only_preflight"),
			operationStep("disable_privileged_login", "external_provisioner"),
			operationStep("clear_privileged_password", "external_provisioner"),
			operationStep("terminate_privileged_sessions", "external_provisioner"),
			operationStep("remove_privileged_memberships", "external_provisioner"),
			operationStep("verify_final_role_catalog", "external_provisioner"),
			operationStep("revoke_mounted_credentials", "external_provisioner"),
			operationStep("publish_cleanup_receipt", "receipt_publication"),
		},
	}
}
