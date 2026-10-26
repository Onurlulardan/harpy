package artifact

import (
	"crypto/ed25519"
	"encoding/base64"
	"time"
)

type verifiedTrustKey struct {
	publicKey   ed25519.PublicKey
	signUntil   time.Time
	verifyUntil time.Time
	active      bool
	revoked     bool
}

type VerifiedTrustBundle struct {
	Kind       RegistryKind
	Generation int64
	SHA256     string
	keys       map[string]verifiedTrustKey
}

func (bundle VerifiedTrustBundle) ActiveKey() (string, ed25519.PublicKey, bool) {
	for kid, key := range bundle.keys {
		if key.active && !key.revoked {
			return kid, append(ed25519.PublicKey(nil), key.publicKey...), true
		}
	}
	return "", nil, false
}

func VerifyTrustBundle(
	document []byte,
	expectedKind RegistryKind,
	now time.Time,
	releaseRoot ed25519.PublicKey,
) (VerifiedTrustBundle, error) {
	var value TrustBundle
	canonical, err := strictCanonicalDecode(document, &value)
	if err != nil || len(releaseRoot) != ed25519.PublicKeySize ||
		value.Schema != TrustBundleSchema || value.RegistryKind != expectedKind ||
		value.TrustGeneration < 1 || !validDigest(value.BundleSHA256) {
		return VerifiedTrustBundle{}, ErrInvalidDocument
	}
	body, err := unsignedTrustBundle(canonical)
	if err != nil || digestHex(body) != value.BundleSHA256 {
		return VerifiedTrustBundle{}, ErrInvalidDocument
	}
	signature, ok := decodeBase64URL(value.ReleaseSignature, ed25519.SignatureSize)
	if !ok || !ed25519.Verify(releaseRoot,
		signatureInput(trustBundleDomain, body), signature) {
		return VerifiedTrustBundle{}, ErrInvalidSignature
	}
	keys, err := parseTrustKeys(value, now, releaseRoot)
	if err != nil {
		return VerifiedTrustBundle{}, err
	}
	return VerifiedTrustBundle{
		Kind: value.RegistryKind, Generation: value.TrustGeneration,
		SHA256: digestHex(canonical), keys: keys,
	}, nil
}

// VerifyPinnedTrustBundle accepts a bundle whose exact canonical digest was
// authenticated by another signed artifact. It still validates every key and
// the release-signature shape, but does not grant unpinned release authority.
func VerifyPinnedTrustBundle(
	document []byte,
	expectedKind RegistryKind,
	expectedSHA256 string,
	now time.Time,
) (VerifiedTrustBundle, error) {
	var value TrustBundle
	canonical, err := strictCanonicalDecode(document, &value)
	if err != nil || digestHex(canonical) != expectedSHA256 ||
		value.Schema != TrustBundleSchema || value.RegistryKind != expectedKind ||
		value.TrustGeneration < 1 || !validDigest(value.BundleSHA256) {
		return VerifiedTrustBundle{}, ErrInvalidDocument
	}
	body, err := unsignedTrustBundle(canonical)
	signature, signatureOK := decodeBase64URL(value.ReleaseSignature, ed25519.SignatureSize)
	if err != nil || digestHex(body) != value.BundleSHA256 || !signatureOK ||
		len(signature) != ed25519.SignatureSize {
		return VerifiedTrustBundle{}, ErrInvalidDocument
	}
	keys, err := parseTrustKeys(value, now, make(ed25519.PublicKey, ed25519.PublicKeySize))
	if err != nil {
		return VerifiedTrustBundle{}, err
	}
	return VerifiedTrustBundle{Kind: value.RegistryKind,
		Generation: value.TrustGeneration, SHA256: expectedSHA256, keys: keys}, nil
}

func parseTrustKeys(
	value TrustBundle,
	now time.Time,
	releaseRoot ed25519.PublicKey,
) (map[string]verifiedTrustKey, error) {
	keys := make(map[string]verifiedTrustKey, len(value.Retired)+1)
	activeKey, notBefore, signUntil, ok := parseActiveKey(value.Active)
	if !ok || now.Before(notBefore) || !now.Before(signUntil) ||
		ed25519.PublicKey(activeKey).Equal(releaseRoot) {
		return nil, ErrUntrustedKey
	}
	keys[value.Active.KID] = verifiedTrustKey{
		publicKey: activeKey, signUntil: signUntil, verifyUntil: signUntil, active: true,
	}
	for _, retired := range value.Retired {
		key, signEnd, verifyEnd, valid := parseRetiredKey(retired)
		if !valid || retired.KID == value.Active.KID || keys[retired.KID].publicKey != nil ||
			ed25519.PublicKey(key).Equal(releaseRoot) {
			return nil, ErrUntrustedKey
		}
		keys[retired.KID] = verifiedTrustKey{
			publicKey: key, signUntil: signEnd, verifyUntil: verifyEnd,
			revoked: retired.Disposition == "revoked",
		}
	}
	return keys, nil
}

func parseActiveKey(value TrustKey) (ed25519.PublicKey, time.Time, time.Time, bool) {
	notBefore, beforeOK := parseInstant(value.NotBefore)
	signUntil, untilOK := parseInstant(value.SignUntil)
	key, keyOK := decodeBase64URL(value.PublicKey, ed25519.PublicKeySize)
	return key, notBefore, signUntil, value.Algorithm == AlgorithmEd25519 &&
		validKID(value.KID) && beforeOK && untilOK && notBefore.Before(signUntil) && keyOK
}

func parseRetiredKey(value RetiredTrustKey) (ed25519.PublicKey, time.Time, time.Time, bool) {
	signUntil, signOK := parseInstant(value.SignUntil)
	verifyUntil, verifyOK := parseInstant(value.VerifyUntil)
	key, keyOK := decodeBase64URL(value.PublicKey, ed25519.PublicKeySize)
	validDisposition := value.Disposition == "verify_only" || value.Disposition == "revoked"
	return key, signUntil, verifyUntil, value.Algorithm == AlgorithmEd25519 &&
		validKID(value.KID) && signOK && verifyOK && !verifyUntil.Before(signUntil) &&
		keyOK && validDisposition
}

func encodeSignature(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}
