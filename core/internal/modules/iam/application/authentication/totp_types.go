package authentication

type TOTPEnvelope struct {
	Profile              string
	CipherSuite          string
	KMSKeyID             string
	Nonce                []byte
	Ciphertext           []byte
	AuthenticationTag    []byte
	AADSHA256            [32]byte
	LastAcceptedTimeStep *int64
	Revision             int64
}

func (envelope TOTPEnvelope) valid() bool {
	return envelope.Profile == "totp_sha256_6_30_v1" &&
		envelope.CipherSuite == "aes_256_gcm_v1" &&
		validKMSKeyID(envelope.KMSKeyID) && len(envelope.Nonce) == 12 &&
		len(envelope.Ciphertext) == 32 && len(envelope.AuthenticationTag) == 16 &&
		envelope.AADSHA256 != [32]byte{} && envelope.Revision > 0 &&
		(envelope.LastAcceptedTimeStep == nil || *envelope.LastAcceptedTimeStep >= 0)
}
