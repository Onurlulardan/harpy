package licenseauthority

import (
	"crypto/ed25519"
	"time"

	licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"
)

type TrustKeyStatus = licensev2.TrustKeyStatus

const (
	TrustKeyActive     = licensev2.TrustKeyActive
	TrustKeyVerifyOnly = licensev2.TrustKeyVerifyOnly
	TrustKeyRevoked    = licensev2.TrustKeyRevoked
)

type trustedLicenseKey struct {
	publicKey            ed25519.PublicKey
	status               TrustKeyStatus
	notBefore, signUntil time.Time
	verifyUntil          time.Time
}

type TrustBundle struct {
	digestHex string
	issuers   map[string]trustedIssuer
}

type trustedIssuer struct {
	sequence int64
	keys     map[string]trustedLicenseKey
}

func (bundle TrustBundle) DigestHex() string { return bundle.digestHex }

func validTrustStatus(value TrustKeyStatus) bool {
	return value == TrustKeyActive || value == TrustKeyVerifyOnly || value == TrustKeyRevoked
}
