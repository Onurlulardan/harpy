package platformauthority

import (
	"crypto/ed25519"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func VerifySaaSTrustBundle(
	document []byte,
	now time.Time,
	releaseRoot ed25519.PublicKey,
) error {
	_, err := ports.VerifyTrustBundle(document, ports.RegistrySaaS, now, releaseRoot)
	return err
}
