package binarycontract

import "github.com/Onurlulardan/harpy/internal/attestation"

func Verify(document []byte) (Artifact, error) {
	verified, err := attestation.VerifyBinarySchemaContract(document)
	if err != nil {
		return Artifact{}, err
	}
	return wrap(verified), nil
}
