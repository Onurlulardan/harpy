package main

import "fmt"

func publishGeneratedEvidence(
	target string, catalog releaseCatalog, signer signingAuthority, evidence generatedEvidence,
	replace bool,
) error {
	output, err := beginOwnedOutput(target, replace)
	if err != nil {
		return err
	}
	defer output.abort()
	artifacts := append(catalog.baseArtifacts(), signer.trustRoot)
	artifacts = append(artifacts,
		evidence.authority.document, evidence.authority.signature,
		evidence.cutoverOperation.document, evidence.cutoverOperation.signature,
		evidence.release.document, evidence.release.signature,
		evidence.bootstrapReceipt, evidence.extensionReceipt, evidence.cleanupReceipt,
		evidence.cutoverReceipt.document, evidence.cutoverReceipt.signature,
	)
	for _, item := range artifacts {
		if err := publishArtifact(output.staging, item); err != nil {
			return fmt.Errorf("publish %s: %w", item.name, err)
		}
	}
	if err := output.commit(); err != nil {
		return err
	}
	return nil
}
