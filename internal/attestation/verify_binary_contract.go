package attestation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func VerifyBinarySchemaContract(document []byte) (VerifiedBinarySchemaContract, error) {
	var contract runtimev1.BinarySchemaContract
	if err := decodeStrict(document, &contract); err != nil {
		return VerifiedBinarySchemaContract{}, fmt.Errorf("decode binary schema contract: %w", err)
	}
	if err := contract.Validate(); err != nil {
		return VerifiedBinarySchemaContract{}, fmt.Errorf("validate binary schema contract: %w", err)
	}
	canonical, err := MarshalCanonical(contract)
	if err != nil {
		return VerifiedBinarySchemaContract{}, fmt.Errorf(
			"canonicalize binary schema contract: %w",
			err,
		)
	}
	if !bytes.Equal(document, canonical) {
		return VerifiedBinarySchemaContract{}, fmt.Errorf(
			"binary schema contract is not exact canonical JSON",
		)
	}
	digest := sha256.Sum256(canonical)
	return VerifiedBinarySchemaContract{
		contract: contract, canonical: canonical,
		digestSHA256: hex.EncodeToString(digest[:]), verified: true,
	}, nil
}
