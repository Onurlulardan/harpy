package binarycontract

import (
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type Artifact struct {
	verified attestation.VerifiedBinarySchemaContract
}

func (artifact Artifact) Value() runtimev1.BinarySchemaContract {
	return artifact.verified.Value()
}

func (artifact Artifact) CanonicalJSON() []byte {
	return artifact.verified.CanonicalJSON()
}

func (artifact Artifact) DigestSHA256() string {
	return artifact.verified.DigestSHA256()
}

func (artifact Artifact) IsVerified() bool {
	return artifact.verified.IsVerified()
}

func wrap(verified attestation.VerifiedBinarySchemaContract) Artifact {
	return Artifact{verified: verified}
}
