package webhookbody

const (
	Algorithm       = "AES-256-GCM"
	MaximumBodySize = 1 << 20
	dekSize         = 32
	nonceSize       = 12
)

type WrappedDEK struct {
	KeyID      string
	Ciphertext []byte
}

type Envelope struct {
	Algorithm  string
	Nonce      []byte
	Ciphertext []byte
	WrappedDEK WrappedDEK
	BodySHA256 [32]byte
}

func (envelope Envelope) valid() bool {
	return envelope.Algorithm == Algorithm && len(envelope.Nonce) == nonceSize &&
		len(envelope.Ciphertext) >= 16 && envelope.WrappedDEK.KeyID != "" &&
		len(envelope.WrappedDEK.Ciphertext) > 0
}
