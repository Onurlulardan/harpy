package contractretirement

import (
	"crypto/sha256"
	"fmt"
	"regexp"
)

const keySealDomain = "harpy/contract-retirement/key-seal/v1\x00"

var keyPattern = regexp.MustCompile(`^rq[.]alias[.][0-9a-f]{64}$`)

type Key struct {
	value string
	seal  [sha256.Size]byte
}

func ParseKey(value string) (Key, error) {
	if !keyPattern.MatchString(value) {
		return Key{}, ErrInvalid
	}
	key := Key{value: value, seal: keySeal(value)}
	return key, nil
}

func (key Key) Validate() error {
	if !keyPattern.MatchString(key.value) || key.seal != keySeal(key.value) {
		return ErrInvalid
	}
	return nil
}

func (key Key) DatabaseValue() (string, error) {
	if err := key.Validate(); err != nil {
		return "", err
	}
	return key.value, nil
}

func keySeal(value string) [sha256.Size]byte {
	return sha256.Sum256([]byte(keySealDomain + value))
}

func (Key) String() string                     { return redaction }
func (Key) GoString() string                   { return redaction }
func (key Key) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Key) MarshalJSON() ([]byte, error)       { return rejectJSON() }
