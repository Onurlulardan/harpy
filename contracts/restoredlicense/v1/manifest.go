package restoredlicensev1

import (
	"crypto/sha256"
	"encoding/hex"
)

type Command string

const ReconcileRestoredLicense Command = "billing.reconcile_restored_license"

func Manifest() []Command { return []Command{ReconcileRestoredLicense} }

func ManifestSHA256() string {
	digest := sha256.Sum256([]byte(string(ReconcileRestoredLicense)))
	return hex.EncodeToString(digest[:])
}
