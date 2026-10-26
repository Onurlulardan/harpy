package attestation

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

type VerifiedBinarySchemaContract struct {
	contract     runtimev1.BinarySchemaContract
	canonical    []byte
	digestSHA256 string
	verified     bool
}

func (contract VerifiedBinarySchemaContract) Value() runtimev1.BinarySchemaContract {
	if !contract.verified {
		return runtimev1.BinarySchemaContract{}
	}
	return contract.contract
}

func (contract VerifiedBinarySchemaContract) CanonicalJSON() []byte {
	if !contract.verified {
		return nil
	}
	return append([]byte(nil), contract.canonical...)
}

func (contract VerifiedBinarySchemaContract) DigestSHA256() string {
	if !contract.verified {
		return ""
	}
	return contract.digestSHA256
}

func (contract VerifiedBinarySchemaContract) IsVerified() bool {
	return contract.verified
}
