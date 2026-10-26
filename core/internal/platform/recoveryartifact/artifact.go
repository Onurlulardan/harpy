package recoveryartifact

import (
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

type TrustBundle = recoverysigningv1.Bundle
type Artifact = environmentrecoveryv1.Artifact

const (
	MaximumTrustBundleBytes = recoverysigningv1.MaximumBundleBytes
	MaximumArtifactBytes    = environmentrecoveryv1.MaximumArtifactBytes
	PurposeRestoreReceipt   = environmentrecoveryv1.PurposeRestoreReceipt
)

func ParseTrustBundle(document []byte) (TrustBundle, error) {
	return recoverysigningv1.Parse(document)
}

func Parse(document []byte, trust TrustBundle, now time.Time) (Artifact, error) {
	return environmentrecoveryv1.Parse(document, trust, now)
}
