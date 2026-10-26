package iam

import (
	iamrestorev1 "github.com/Onurlulardan/harpy/contracts/iamrestore/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/custodymanifest"
)

type RestoredSecurityEvidence = iamrestorev1.SecurityEvidence
type RestoreCustodyManifest = custodymanifest.Manifest
type RestoreCustodyResult = custodymanifest.Result

const MaximumRestoreDocumentBytes = iamrestorev1.MaximumDocumentBytes

func ParseRestoredSecurityEvidence(document []byte) (RestoredSecurityEvidence, error) {
	return iamrestorev1.ParseSecurityEvidence(document)
}

func ParseRestoreCustodyManifest(
	document []byte,
	environmentID UUID,
) (RestoreCustodyManifest, error) {
	return custodymanifest.Parse(document, environmentID)
}
