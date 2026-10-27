package ports

import (
	"crypto/ed25519"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

type RegistryKind = artifact.RegistryKind
type VerifiedTrustBundle = artifact.VerifiedTrustBundle

const (
	RegistryDevelopment = artifact.RegistryDevelopment
	RegistrySaaS        = artifact.RegistrySaaS
)

func VerifyTrustBundle(
	document []byte,
	expectedKind RegistryKind,
	now time.Time,
	releaseRoot ed25519.PublicKey,
) (VerifiedTrustBundle, error) {
	return artifact.VerifyTrustBundle(document, expectedKind, now, releaseRoot)
}
