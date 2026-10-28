package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type releaseFile struct {
	key  string
	name string
	body []byte
}

func releaseMaterial(root, profile string) ([]string, error) {
	if err := validProfile(profile); err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	directory := filepath.Join(absolute, "deploy", profile, "secrets", "release")
	if !setupRotateSecrets && !setupRefreshRelease {
		lines, ok, reuseErr := reuseCompleteRelease(directory)
		if reuseErr != nil || ok {
			return lines, reuseErr
		}
	}
	if err = os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	fresh := newReleaseClock(time.Now())
	source, err := preservedSourceClock(directory, fresh)
	if err != nil {
		return nil, err
	}
	files, err := buildReleaseBundle(fresh, source)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(files))
	for _, file := range files {
		path := filepath.Join(directory, file.name)
		if err = writePrivate(path, file.body); err != nil {
			return nil, fmt.Errorf("write %s: %w", file.name, err)
		}
		lines = append(lines, file.key+"="+path)
	}
	return lines, nil
}

func buildReleaseFiles(clock releaseClock) ([]releaseFile, error) {
	return buildReleaseBundle(clock, clock)
}

func buildReleaseBundle(fresh, source releaseClock) ([]releaseFile, error) {
	keys, err := newReleaseKeys(fresh)
	if err != nil {
		return nil, err
	}
	keys, err = coverInstalledIssuedAt(keys, source)
	if err != nil {
		return nil, err
	}
	sourceKeys := keys
	sourceKeys.clock = source
	manifests, err := buildPostgresManifests()
	if err != nil {
		return nil, fmt.Errorf("postgres manifests: %w", err)
	}
	if err = applySchemaContractManifests(&manifests); err != nil {
		return nil, err
	}
	operations, err := buildPostgresOperations(manifests)
	if err != nil {
		return nil, fmt.Errorf("postgres operations: %w", err)
	}
	authority, authoritySignature, err := buildAuthority(sourceKeys, manifests, operations)
	if err != nil {
		return nil, fmt.Errorf("bootstrap authority: %w", err)
	}
	verifiedAuthority, err := attestation.VerifyBootstrapAuthority(
		authority.body, authoritySignature, keys.root, source.now,
	)
	if err != nil {
		return nil, err
	}
	catalog, err := buildCatalogSet(keys)
	if err != nil {
		return nil, fmt.Errorf("catalog package set: %w", err)
	}
	operationDocument, operationSignature, err := buildCutoverOperation(
		sourceKeys, manifests.migration.digest, authority.digest,
	)
	if err != nil {
		return nil, fmt.Errorf("cutover authority operation: %w", err)
	}
	releaseDocument, releaseSignature, verifiedRelease, err := buildReleaseManifest(
		keys, manifests, operations, authority.digest, catalogSetDigest(catalog), operationDocument.digest,
	)
	if err != nil {
		return nil, fmt.Errorf("release manifest: %w", err)
	}
	if err = verifyCatalogSet(keys, catalog, verifiedRelease); err != nil {
		return nil, fmt.Errorf("catalog package set: %w", err)
	}
	verifiedOperation, err := verifyCutoverOperation(
		sourceKeys, operationDocument.body, operationSignature, verifiedRelease, verifiedAuthority,
	)
	if err != nil {
		return nil, fmt.Errorf("cutover authority operation: %w", err)
	}
	var operation cutoverv1.AuthorityAdoptionOperation
	if err = attestation.DecodeStrict(operationDocument.body, &operation); err != nil {
		return nil, err
	}
	bootstrap, extensionReceipt, cleanup, err := buildRuntimeReceipts(
		keys, manifests, operations, authority, releaseDocument,
	)
	if err != nil {
		return nil, fmt.Errorf("runtime receipts: %w", err)
	}
	receiptKeys := keys
	receiptKeys.clock.adopted = source.adopted
	cutoverReceipt, cutoverReceiptSignature, err := buildCutoverReceipt(
		receiptKeys, operation, verifiedOperation, verifiedRelease, verifiedAuthority,
	)
	if err != nil {
		return nil, fmt.Errorf("cutover authority receipt: %w", err)
	}
	return bindReleaseBodies(
		releaseDocument.body, releaseSignature, keys.rootDocument, bootstrap, extensionReceipt,
		authority.body, authoritySignature, cleanup, manifests.migration.body, manifests.extension.body,
		manifests.roles.body, manifests.privileges.body, manifests.ownership.body,
		operations.bootstrap.body, operations.extension.body, operations.cleanup.body,
		operationDocument.body, operationSignature, cutoverReceipt, cutoverReceiptSignature,
		catalog, keys.keyring,
	)
}
