package main

import (
	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type postgresOperations struct {
	bootstrap, extension, cleanup releaseDocument
}

func buildPostgresOperations(manifests postgresManifests) (postgresOperations, error) {
	bootstrap, err := canonicalDocument(postgresv1.BootstrapOperation{
		ContractVersion: "postgres-bootstrap-operation.v1", DatabaseOrigin: "fresh",
		DatabaseName: localDatabase,
		InputHashes: postgresv1.BootstrapInputHashes{
			RolesManifestSHA256:      manifests.roles.digest,
			PrivilegesManifestSHA256: manifests.privileges.digest,
			OwnershipManifestSHA256:  manifests.ownership.digest,
		},
		Steps: operationSteps(
			"verify_artifact_hashes", "read_only_preflight",
			"verify_credential_lease", "read_only_preflight",
			"verify_external_cutover_authority_owner", "read_only_preflight",
			"create_database", "autocommit_ddl",
			"reconcile_roles", "transactional_ddl",
			"reconcile_memberships", "transactional_ddl",
			"reconcile_privileges", "transactional_ddl",
			"verify_ownership_inventory", "read_only_preflight",
			"verify_final_catalog", "read_only_preflight",
			"publish_bootstrap_receipt", "receipt_publication",
		),
	})
	if err != nil {
		return postgresOperations{}, err
	}
	if _, err = attestation.VerifyBootstrapOperationArtifact(bootstrap.body); err != nil {
		return postgresOperations{}, err
	}
	extension, err := canonicalDocument(postgresv1.ExtensionOperation{
		ContractVersion: "postgres-extension-operation.v1", DatabaseOrigin: "fresh",
		DatabaseName: localDatabase,
		InputHashes: postgresv1.ExtensionInputHashes{
			ExtensionManifestSHA256: manifests.extension.digest,
			OwnershipManifestSHA256: manifests.ownership.digest,
		},
		Steps: operationSteps(
			"verify_artifact_hashes", "read_only_preflight",
			"verify_credential_lease", "read_only_preflight",
			"verify_bootstrap_receipt", "read_only_preflight",
			"resolve_control_files", "read_only_preflight",
			"reconcile_extensions", "transactional_ddl",
			"verify_extension_catalog", "read_only_preflight",
			"verify_ownership_inventory", "read_only_preflight",
			"publish_extension_receipt", "receipt_publication",
		),
	})
	if err != nil {
		return postgresOperations{}, err
	}
	if _, err = attestation.VerifyExtensionOperationArtifact(extension.body); err != nil {
		return postgresOperations{}, err
	}
	cleanup, err := canonicalDocument(postgresv1.CredentialCleanupOperation{
		ContractVersion: "credential-cleanup-operation.v1", DatabaseOrigin: "fresh",
		DatabaseName: localDatabase,
		InputHashes: postgresv1.CredentialCleanupHashes{
			RolesManifestSHA256:      manifests.roles.digest,
			PrivilegesManifestSHA256: manifests.privileges.digest,
		},
		PrivilegedRoles: []string{"harpy_cluster_admin", "harpy_extension_admin"},
		Steps: operationSteps(
			"verify_artifact_hashes", "read_only_preflight",
			"disable_privileged_login", "external_provisioner",
			"clear_privileged_password", "external_provisioner",
			"terminate_privileged_sessions", "external_provisioner",
			"remove_privileged_memberships", "external_provisioner",
			"verify_final_role_catalog", "external_provisioner",
			"revoke_mounted_credentials", "external_provisioner",
			"publish_cleanup_receipt", "receipt_publication",
		),
	})
	if err != nil {
		return postgresOperations{}, err
	}
	if _, err = attestation.VerifyCredentialCleanupOperationArtifact(cleanup.body); err != nil {
		return postgresOperations{}, err
	}
	return postgresOperations{bootstrap: bootstrap, extension: extension, cleanup: cleanup}, nil
}

func operationSteps(pairs ...string) []postgresv1.OperationStep {
	steps := make([]postgresv1.OperationStep, 0, len(pairs)/2)
	for index := 0; index < len(pairs); index += 2 {
		steps = append(steps, postgresv1.OperationStep{
			StepID: pairs[index], TransactionClass: pairs[index+1],
		})
	}
	return steps
}
