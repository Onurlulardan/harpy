package config

type LicenseSourceKind string

const (
	LicenseSourceConnected LicenseSourceKind = "connected"
	LicenseSourceSnapshot  LicenseSourceKind = "snapshot"
)

type LicenseSource struct {
	kind       LicenseSourceKind
	credential ArtifactPath
	issuer     ServiceURL
	snapshot   ArtifactPath
}

func (source LicenseSource) Kind() LicenseSourceKind      { return source.kind }
func (source LicenseSource) CredentialFile() ArtifactPath { return source.credential }
func (source LicenseSource) IssuerURL() ServiceURL        { return source.issuer }
func (source LicenseSource) SnapshotFile() ArtifactPath   { return source.snapshot }

func loadLicenseSource(
	values environmentValues,
	environment Environment,
) (LicenseSource, error) {
	_, credential := values.lookup(licenseCredentialKey)
	_, issuer := values.lookup(licenseIssuerURLKey)
	_, snapshot := values.lookup(licenseSnapshotKey)
	connected := credential || issuer
	if connected == snapshot {
		return LicenseSource{}, configError(
			"HARPY_LICENSE_*", "connected pair and snapshot must form an exact XOR",
		)
	}
	if connected {
		if !credential || !issuer {
			return LicenseSource{}, configError(
				"HARPY_LICENSE_*", "connected source requires credential and issuer",
			)
		}
		credentialPath, err := loadReadArtifact(values, licenseCredentialKey, true)
		if err != nil {
			return LicenseSource{}, err
		}
		issuerURL, err := loadServiceURL(values, licenseIssuerURLKey, environment)
		if err != nil {
			return LicenseSource{}, err
		}
		return LicenseSource{
			kind: LicenseSourceConnected, credential: credentialPath, issuer: issuerURL,
		}, nil
	}
	snapshotPath, err := loadReadArtifact(values, licenseSnapshotKey, false)
	if err != nil {
		return LicenseSource{}, err
	}
	return LicenseSource{kind: LicenseSourceSnapshot, snapshot: snapshotPath}, nil
}

func loadLicenseTrust(values environmentValues) (ArtifactPath, error) {
	return loadReadArtifact(values, licenseTrustBundleKey, false)
}

func loadRecoveryTrust(values environmentValues) (ArtifactPath, error) {
	return loadReadArtifact(values, recoveryTrustBundleKey, false)
}

func loadAuthorityAckTrust(values environmentValues) (ArtifactPath, error) {
	return loadReadArtifact(values, ackTrustBundleKey, false)
}
