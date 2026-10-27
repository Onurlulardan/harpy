package platformauthority

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func SelectSaaSReleaseRoot(
	document []byte,
	ackTrust []byte,
	now time.Time,
) (ed25519.PublicKey, error) {
	root, err := attestation.ParseTrustRoot(document)
	if err != nil {
		return nil, fmt.Errorf("parse release trust root: %w", err)
	}
	var selected ed25519.PublicKey
	for _, candidate := range root.ActivePublicKeys(now) {
		if verifyErr := VerifySaaSTrustBundle(ackTrust, now, candidate); verifyErr == nil {
			if selected != nil {
				return nil, fmt.Errorf("authority trust has ambiguous release root")
			}
			selected = append(ed25519.PublicKey(nil), candidate...)
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("authority trust lacks a valid release root")
	}
	return selected, nil
}
