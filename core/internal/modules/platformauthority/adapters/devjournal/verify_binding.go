package devjournal

import (
	"crypto/ed25519"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func verifyBinding(
	document []byte,
	key ed25519.PublicKey,
) (artifact.VerifiedBindingProof, error) {
	proof, err := artifact.VerifyBindingProof(document, key)
	if err != nil {
		return artifact.VerifiedBindingProof{}, fmt.Errorf("development authority binding proof is invalid")
	}
	return proof, nil
}
