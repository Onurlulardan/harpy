package fcartv1

import (
	"fmt"
	"regexp"
)

var keyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type KeyEnvelope struct {
	KEKID      string
	Ciphertext []byte
	Nonce      [NonceBytes]byte
}

func (envelope KeyEnvelope) Validate() error {
	if !validKeyID(envelope.KEKID) || len(envelope.Ciphertext) != DEKBytes+TagBytes ||
		envelope.Nonce == [NonceBytes]byte{} {
		return ErrInvalidKeyEnvelope
	}
	var nonzero byte
	for _, value := range envelope.Ciphertext {
		nonzero |= value
	}
	if nonzero == 0 {
		return ErrInvalidKeyEnvelope
	}
	return nil
}

func (envelope KeyEnvelope) OwnedCopy() KeyEnvelope {
	envelope.Ciphertext = append([]byte(nil), envelope.Ciphertext...)
	return envelope
}

func (KeyEnvelope) String() string   { return "[REDACTED FCART1 KEY ENVELOPE]" }
func (KeyEnvelope) GoString() string { return "[REDACTED FCART1 KEY ENVELOPE]" }
func (KeyEnvelope) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("[REDACTED FCART1 KEY ENVELOPE]"))
}

func validKeyID(keyID string) bool { return keyIDPattern.MatchString(keyID) }
