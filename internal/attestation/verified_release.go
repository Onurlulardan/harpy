package attestation

import (
	"time"
)

type VerifiedRelease struct {
	manifest       ReleaseManifest
	manifestSHA256 string
	verified       bool
}

func newVerifiedRelease(manifest ReleaseManifest, manifestSHA256 string) VerifiedRelease {
	return VerifiedRelease{
		manifest: manifest, manifestSHA256: manifestSHA256, verified: true,
	}
}

func (release VerifiedRelease) Value() ReleaseManifest {
	if !release.verified {
		return ReleaseManifest{}
	}
	return release.manifest
}

func (release VerifiedRelease) ManifestSHA256() string {
	if !release.verified {
		return ""
	}
	return release.manifestSHA256
}

// IssuedAt returns the timestamp authenticated by the release signature.
func (release VerifiedRelease) IssuedAt() time.Time {
	if !release.verified {
		return time.Time{}
	}
	issuedAt, err := parseUTCTimestamp(release.manifest.IssuedAt)
	if err != nil {
		return time.Time{}
	}
	return issuedAt
}
