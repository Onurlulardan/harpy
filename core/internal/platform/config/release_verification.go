package config

type ReleaseVerification struct {
	inputs            ReleaseInputs
	authority         AuthorityArtifacts
	bootstrapReceipt  AbsolutePath
	extensionReceipt  AbsolutePath
	cleanupReceipt    AbsolutePath
}

func loadReleaseVerification(values environmentValues) (ReleaseVerification, error) {
	inputs, err := loadReleaseInputs(values)
	if err != nil {
		return ReleaseVerification{}, err
	}
	authority, err := loadAuthorityArtifacts(values)
	if err != nil {
		return ReleaseVerification{}, err
	}
	bootstrap, err := loadAbsolutePath(values, "HARPY_BOOTSTRAP_RECEIPT_PATH")
	if err != nil {
		return ReleaseVerification{}, err
	}
	extension, err := loadAbsolutePath(values, "HARPY_EXTENSION_RECEIPT_PATH")
	if err != nil {
		return ReleaseVerification{}, err
	}
	cleanup, err := loadAbsolutePath(values, "HARPY_CREDENTIAL_CLEANUP_RECEIPT_PATH")
	if err != nil {
		return ReleaseVerification{}, err
	}
	if duplicateReleasePath(
		inputs.ManifestPath(), inputs.SignaturePath(), inputs.TrustRootPath(),
		authority.ManifestPath(), authority.SignaturePath(), bootstrap, extension, cleanup,
	) {
		return ReleaseVerification{}, configError(
			"HARPY_*_RECEIPT_PATH", "verification paths must be distinct",
		)
	}
	return ReleaseVerification{
		inputs: inputs, authority: authority, bootstrapReceipt: bootstrap,
		extensionReceipt: extension, cleanupReceipt: cleanup,
	}, nil
}

func (verification ReleaseVerification) Inputs() ReleaseInputs {
	return verification.inputs
}

func (verification ReleaseVerification) Authority() AuthorityArtifacts {
	return verification.authority
}

func (verification ReleaseVerification) BootstrapReceiptPath() AbsolutePath {
	return verification.bootstrapReceipt
}

func (verification ReleaseVerification) ExtensionReceiptPath() AbsolutePath {
	return verification.extensionReceipt
}

func (verification ReleaseVerification) CleanupReceiptPath() AbsolutePath {
	return verification.cleanupReceipt
}

func (verification ReleaseVerification) paths() []AbsolutePath {
	return []AbsolutePath{
		verification.inputs.ManifestPath(), verification.inputs.SignaturePath(),
		verification.inputs.TrustRootPath(), verification.authority.ManifestPath(),
		verification.authority.SignaturePath(), verification.bootstrapReceipt,
		verification.extensionReceipt, verification.cleanupReceipt,
	}
}
