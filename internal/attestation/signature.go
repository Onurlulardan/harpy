package attestation

const algorithmEd25519 = "Ed25519"

type Signature struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Value     string `json:"value"`
}

type trustRootDocument struct {
	ContractVersion string             `json:"contract_version"`
	Keys            []trustedKeyRecord `json:"keys"`
}

type trustedKeyRecord struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
	NotBefore string `json:"not_before"`
	NotAfter  string `json:"not_after"`
}

type signingKeyDocument struct {
	ContractVersion string `json:"contract_version"`
	KeyID           string `json:"key_id"`
	Algorithm       string `json:"algorithm"`
	Seed            string `json:"seed"`
}
