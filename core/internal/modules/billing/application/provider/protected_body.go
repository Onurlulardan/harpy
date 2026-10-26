package providerapp

const maximumWebhookBodyBytes = 1 << 20

const (
	maximumEnvelopeOverheadBytes = 64
	maximumWrappedDEKBytes       = 8 << 10
	maximumKeyIDBytes            = 256
	maximumAlgorithmBytes        = 64
)

type ProtectedBody struct {
	algorithm  string
	nonce      []byte
	ciphertext []byte
	keyID      string
	wrappedDEK []byte
	bodySHA256 [32]byte
}

func NewProtectedBody(
	algorithm string,
	nonce []byte,
	ciphertext []byte,
	keyID string,
	wrappedDEK []byte,
	bodySHA256 [32]byte,
) (ProtectedBody, error) {
	if algorithm == "" || len(algorithm) > maximumAlgorithmBytes ||
		len(nonce) == 0 || len(nonce) > maximumEnvelopeOverheadBytes ||
		len(ciphertext) == 0 ||
		len(ciphertext) > maximumWebhookBodyBytes+maximumEnvelopeOverheadBytes ||
		keyID == "" || len(keyID) > maximumKeyIDBytes ||
		len(wrappedDEK) == 0 || len(wrappedDEK) > maximumWrappedDEKBytes ||
		bodySHA256 == [32]byte{} {
		return ProtectedBody{}, ErrWebhookRequest
	}
	return ProtectedBody{
		algorithm: algorithm, nonce: append([]byte(nil), nonce...),
		ciphertext: append([]byte(nil), ciphertext...), keyID: keyID,
		wrappedDEK: append([]byte(nil), wrappedDEK...), bodySHA256: bodySHA256,
	}, nil
}

func (body ProtectedBody) Algorithm() string    { return body.algorithm }
func (body ProtectedBody) Nonce() []byte        { return append([]byte(nil), body.nonce...) }
func (body ProtectedBody) Ciphertext() []byte   { return append([]byte(nil), body.ciphertext...) }
func (body ProtectedBody) KeyID() string        { return body.keyID }
func (body ProtectedBody) WrappedDEK() []byte   { return append([]byte(nil), body.wrappedDEK...) }
func (body ProtectedBody) BodySHA256() [32]byte { return body.bodySHA256 }

func (body ProtectedBody) valid() bool {
	validated, err := NewProtectedBody(
		body.algorithm, body.nonce, body.ciphertext, body.keyID,
		body.wrappedDEK, body.bodySHA256,
	)
	return err == nil && protectedBodiesEqual(validated, body)
}

func protectedBodiesEqual(left, right ProtectedBody) bool {
	if left.algorithm != right.algorithm || left.keyID != right.keyID ||
		left.bodySHA256 != right.bodySHA256 {
		return false
	}
	return string(left.nonce) == string(right.nonce) &&
		string(left.ciphertext) == string(right.ciphertext) &&
		string(left.wrappedDEK) == string(right.wrappedDEK)
}
