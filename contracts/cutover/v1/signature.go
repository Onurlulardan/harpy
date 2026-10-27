package cutoverv1

type Signature struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Value     string `json:"value"`
}

type CredentialLeaseSignature struct {
	ContractVersion string    `json:"contract_version"`
	ManifestSHA256  string    `json:"manifest_sha256"`
	Signature       Signature `json:"signature"`
}

func ValidateCredentialLeaseSignature(value CredentialLeaseSignature) error {
	if value.ContractVersion != "cutover-credential-lease-signature.v1" ||
		!validDigest(value.ManifestSHA256) {
		return contractError("credential lease signature contract")
	}
	return validateSignature(value.Signature)
}
