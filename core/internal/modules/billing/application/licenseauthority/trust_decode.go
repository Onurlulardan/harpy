package licenseauthority

import (
	"crypto/ed25519"
	"encoding/base64"
	"time"
)

func parseTrustKey(
	keyID, algorithm, encoded string,
	status TrustKeyStatus,
	notBeforeValue, signUntilValue, verifyUntilValue string,
) (trustedLicenseKey, error) {
	if !keyIDPattern.MatchString(keyID) || algorithm != SignatureAlgorithm ||
		!validTrustStatus(status) {
		return trustedLicenseKey{}, ErrInvalidTrustBundle
	}
	publicKey, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	if err != nil || len(publicKey) != ed25519.PublicKeySize ||
		base64.RawURLEncoding.EncodeToString(publicKey) != encoded {
		return trustedLicenseKey{}, ErrInvalidTrustBundle
	}
	notBefore, valid := parseExactInstant(notBeforeValue)
	if !valid {
		return trustedLicenseKey{}, ErrInvalidTrustBundle
	}
	signUntil, valid := parseExactInstant(signUntilValue)
	if !valid {
		return trustedLicenseKey{}, ErrInvalidTrustBundle
	}
	verifyUntil, valid := parseExactInstant(verifyUntilValue)
	if !valid || !notBefore.Before(signUntil) || signUntil.After(verifyUntil) {
		return trustedLicenseKey{}, ErrInvalidTrustBundle
	}
	return trustedLicenseKey{
		publicKey: ed25519.PublicKey(publicKey), status: status,
		notBefore: notBefore, signUntil: signUntil, verifyUntil: verifyUntil,
	}, nil
}

func keyCanVerify(key trustedLicenseKey, issuedAt, verifiedAt time.Time) bool {
	if key.status == TrustKeyRevoked || issuedAt.Before(key.notBefore) ||
		issuedAt.After(key.signUntil) || verifiedAt.After(key.verifyUntil) {
		return false
	}
	return key.status == TrustKeyActive || key.status == TrustKeyVerifyOnly
}
