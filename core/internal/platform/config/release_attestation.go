package config

type ReleaseAttestation struct {
	inputs      ReleaseInputs
	authority   PrivilegedAuthority
	receiptPath AbsolutePath
	signingKey  Secret
}

func loadReleaseAttestation(
	values environmentValues,
	receiptKey string,
) (ReleaseAttestation, error) {
	inputs, err := loadReleaseInputs(values)
	if err != nil {
		return ReleaseAttestation{}, err
	}
	authority, err := loadPrivilegedAuthority(values)
	if err != nil {
		return ReleaseAttestation{}, err
	}
	receipt, err := loadAbsolutePath(values, receiptKey)
	if err != nil {
		return ReleaseAttestation{}, err
	}
	signingKey, err := loadSecret(
		values, "HARPY_ATTESTATION_SIGNING_KEY", true, secretFileLimit,
	)
	if err != nil {
		return ReleaseAttestation{}, err
	}
	if duplicateReleasePath(
		inputs.ManifestPath(), inputs.SignaturePath(), inputs.TrustRootPath(),
		authority.Artifacts().ManifestPath(), authority.Artifacts().SignaturePath(), receipt,
	) {
		return ReleaseAttestation{}, configError(receiptKey, "must not reuse a release input path")
	}
	return ReleaseAttestation{
		inputs: inputs, authority: authority, receiptPath: receipt, signingKey: signingKey,
	}, nil
}

func (release ReleaseAttestation) Inputs() ReleaseInputs       { return release.inputs }
func (release ReleaseAttestation) ManifestPath() AbsolutePath  { return release.inputs.ManifestPath() }
func (release ReleaseAttestation) SignaturePath() AbsolutePath { return release.inputs.SignaturePath() }
func (release ReleaseAttestation) TrustRootPath() AbsolutePath { return release.inputs.TrustRootPath() }
func (release ReleaseAttestation) Authority() PrivilegedAuthority {
	return release.authority
}
func (release ReleaseAttestation) ReceiptPath() AbsolutePath { return release.receiptPath }
func (release ReleaseAttestation) SigningKey() Secret        { return release.signingKey }

func (release ReleaseAttestation) paths() []AbsolutePath {
	return []AbsolutePath{
		release.inputs.ManifestPath(), release.inputs.SignaturePath(), release.inputs.TrustRootPath(),
		release.authority.Artifacts().ManifestPath(), release.authority.Artifacts().SignaturePath(),
		release.receiptPath,
	}
}

func duplicateReleasePath(paths ...AbsolutePath) bool {
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if _, exists := seen[path.Value()]; exists {
			return true
		}
		seen[path.Value()] = struct{}{}
	}
	return false
}
