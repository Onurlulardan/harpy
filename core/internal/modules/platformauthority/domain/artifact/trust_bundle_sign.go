package artifact

import (
	"crypto/ed25519"
)

// SignTrustBundle is used by the release pipeline, never by an authority registry.
func SignTrustBundle(
	value TrustBundle,
	releasePrivateKey ed25519.PrivateKey,
) ([]byte, error) {
	if len(releasePrivateKey) != ed25519.PrivateKeySize ||
		value.BundleSHA256 != "" || value.ReleaseSignature != "" {
		return nil, ErrInvalidDocument
	}
	placeholder, err := canonicalValue(value)
	if err != nil {
		return nil, err
	}
	body, err := unsignedTrustBundle(placeholder)
	if err != nil {
		return nil, err
	}
	value.BundleSHA256 = digestHex(body)
	value.ReleaseSignature = encodeSignature(ed25519.Sign(
		releasePrivateKey, signatureInput(trustBundleDomain, body),
	))
	return canonicalValue(value)
}
