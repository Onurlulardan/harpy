package attestation

func DecodeStrict(document []byte, target any) error {
	return decodeStrict(document, target)
}
