package fcartv1

import (
	"crypto/sha256"
	"encoding/binary"
)

const envelopeDigestDomain = "FCART1-DEK-ENVELOPE.v1"

func (envelope KeyEnvelope) SHA256() ([sha256.Size]byte, error) {
	if err := envelope.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(envelopeDigestDomain))
	var keyIDLength [2]byte
	binary.BigEndian.PutUint16(keyIDLength[:], uint16(len(envelope.KEKID)))
	_, _ = hash.Write(keyIDLength[:])
	_, _ = hash.Write([]byte(envelope.KEKID))
	_, _ = hash.Write(envelope.Nonce[:])
	_, _ = hash.Write(envelope.Ciphertext)
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}
