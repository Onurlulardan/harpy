package config

type ReleaseInputs struct {
	manifestPath  AbsolutePath
	signaturePath AbsolutePath
	trustRootPath AbsolutePath
}

func loadReleaseInputs(values environmentValues) (ReleaseInputs, error) {
	manifest, err := loadAbsolutePath(values, "HARPY_RELEASE_MANIFEST_PATH")
	if err != nil {
		return ReleaseInputs{}, err
	}
	signature, err := loadAbsolutePath(values, "HARPY_RELEASE_SIGNATURE_PATH")
	if err != nil {
		return ReleaseInputs{}, err
	}
	trustRoot, err := loadAbsolutePath(values, "HARPY_RELEASE_TRUST_ROOT_PATH")
	if err != nil {
		return ReleaseInputs{}, err
	}
	if duplicateReleasePath(manifest, signature, trustRoot) {
		return ReleaseInputs{}, configError(
			"HARPY_RELEASE_*_PATH", "release input paths must be distinct",
		)
	}
	return ReleaseInputs{
		manifestPath: manifest, signaturePath: signature, trustRootPath: trustRoot,
	}, nil
}

func (inputs ReleaseInputs) ManifestPath() AbsolutePath  { return inputs.manifestPath }
func (inputs ReleaseInputs) SignaturePath() AbsolutePath { return inputs.signaturePath }
func (inputs ReleaseInputs) TrustRootPath() AbsolutePath { return inputs.trustRootPath }
