//go:build integration

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type workflowAuthorityFixture struct {
	paths         authorityArtifactPaths
	signer        workflowSigner
	operationPath string
	rootDSNPath   string
	receiptPath   string
	signaturePath string
}

func prepareWorkflowAuthorityArtifacts(
	t *testing.T, ctx context.Context, database workflowDatabase,
	directory string, now time.Time,
) workflowAuthorityFixture {
	t.Helper()
	signer := prepareWorkflowSigner(t, directory, now)
	bootstrap := workflowBootstrapAuthority(now)
	if err := runtimev1.ValidateBootstrapAuthorityManifest(bootstrap); err != nil {
		t.Fatalf("validate bootstrap authority fixture: %v", err)
	}
	bootstrapPath, bootstrapDocument := writeWorkflowArtifact(
		t, directory, "bootstrap-authority.json", bootstrap,
	)
	bootstrapSignature := signWorkflowDocument(t, signer, bootstrapDocument,
		now.Add(-time.Minute), "bootstrap-authority-signature.v1")
	bootstrapSignaturePath := writeWorkflowBytes(
		t, directory, "bootstrap-authority.sig", bootstrapSignature,
	)
	bootstrapSHA, err := attestation.DigestHex(bootstrapDocument)
	if err != nil {
		t.Fatal(err)
	}
	migrationSHA := strings.Repeat("b", 64)
	operation := buildWorkflowAuthorityOperation(
		t, ctx, database, now, bootstrapSHA, migrationSHA,
	)
	operationPath, operationDocument := writeWorkflowArtifact(
		t, directory, "authority-operation.json", operation,
	)
	operationSignature := signWorkflowDocument(t, signer, operationDocument,
		now.Add(-time.Minute), "cutover-authority-adoption-operation-signature.v1")
	operationSignaturePath := writeWorkflowBytes(
		t, directory, "authority-operation.sig", operationSignature,
	)
	operationSHA, err := attestation.DigestHex(operationDocument)
	if err != nil {
		t.Fatal(err)
	}
	release := workflowReleaseManifest(now, bootstrapSHA, operationSHA, migrationSHA)
	releasePath, releaseDocument := writeWorkflowArtifact(
		t, directory, "release.json", release,
	)
	releaseSignature := signWorkflowDocument(
		t, signer, releaseDocument, now.Add(-time.Minute), "release-signature.v1",
	)
	releaseSignaturePath := writeWorkflowBytes(
		t, directory, "release.sig", releaseSignature,
	)
	rootDSNPath := writeWorkflowBytes(t, directory, "root.dsn", []byte(database.dsn))
	return workflowAuthorityFixture{signer: signer, operationPath: operationPath,
		rootDSNPath: rootDSNPath, receiptPath: directory + "/authority-receipt.json",
		signaturePath: directory + "/authority-receipt.sig", paths: authorityArtifactPaths{
			TrustRoot: signer.rootPath, ReleaseManifest: releasePath,
			ReleaseSignature: releaseSignaturePath, BootstrapAuthority: bootstrapPath,
			BootstrapAuthoritySignature: bootstrapSignaturePath,
			AuthorityOperation:          operationPath,
			AuthorityOperationSignature: operationSignaturePath,
		},
	}
}

func workflowReleaseManifest(
	now time.Time, bootstrapSHA, operationSHA, migrationSHA string,
) attestation.ReleaseManifest {
	digest := strings.Repeat("c", 64)
	return attestation.ReleaseManifest{
		ContractVersion: "postgres-release-manifest.v1", ReleaseID: workflowReleaseID,
		IssuedAt:     now.Add(-time.Minute).Format(time.RFC3339),
		ExpiresAt:    now.Add(10 * time.Minute).Format(time.RFC3339),
		DatabaseName: workflowDatabaseName, DatabaseOrigin: "fresh",
		DatabaseEncoding:       runtimev1.PostgresEncoding,
		DatabaseLocaleProvider: runtimev1.PostgresLocaleProvider,
		DatabaseCollation:      runtimev1.PostgresCollation,
		DatabaseCType:          runtimev1.PostgresCharacterType,
		DatabaseOwner:          runtimev1.DatabaseOwnerRole,
		DataChecksums:          runtimev1.PostgresDataChecksums,
		PostgresImageDigest:    runtimev1.PostgresImageDigest,
		PostgresEngineVersion:  runtimev1.PostgresEngineVersion,
		VectorExtensionVersion: runtimev1.VectorExtensionVersion,
		RuntimeGeneration:      1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: migrationSHA, ExtensionManifestSHA256: digest,
		RolesManifestSHA256: digest, PrivilegesManifestSHA256: digest,
		OwnershipManifestSHA256: digest, BootstrapAuthoritySHA256: bootstrapSHA,
		BootstrapOperationDigest: digest, ExtensionOperationDigest: digest,
		CredentialCleanupDigest: digest, CutoverAuthorityOperationDigest: operationSHA,
		CatalogPackageSetSHA256: digest,
	}
}

func (value workflowAuthorityFixture) adoptionInvocation() invocation {
	return invocation{ContractVersion: invocationContract, Action: "adopt_authority",
		Environment: "test", DeadlineSeconds: 120, Adopt: &adoptionInvocation{
			RootDSNFile: value.rootDSNPath, authorityArtifactPaths: value.paths,
			SigningKeyFile: value.signer.signerPath, ReceiptPath: value.receiptPath,
			ReceiptSignaturePath: value.signaturePath,
		}}
}
