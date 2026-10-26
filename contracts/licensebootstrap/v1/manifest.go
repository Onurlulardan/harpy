package licensebootstrapv1

import (
	"crypto/sha256"
	"encoding/hex"
)

type Command string

const PublishLicenseDerivedOffering Command = "billing.publish_license_derived_offering"

func Manifest() []Command { return []Command{PublishLicenseDerivedOffering} }

func ManifestSHA256() string {
	digest := sha256.Sum256([]byte(string(PublishLicenseDerivedOffering)))
	return hex.EncodeToString(digest[:])
}
