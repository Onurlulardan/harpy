package v1query

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	retirementKeyPrefix  = "rq.alias."
	retirementKeyDomain  = "forgecrud/rest-query/v1-alias-retirement/v1\x00"
	retirementSealDomain = "forgecrud/rest-query/v1-alias-retirement-key-seal/v1\x00"
)

// RetirementKey is the opaque persisted identity of one manifest-alias pair.
type RetirementKey struct {
	value     string
	stateSeal [sha256.Size]byte
	valid     bool
}

func NewRetirementKey(manifestID string, aliasID string) (RetirementKey, error) {
	if !validManifestID(manifestID) || !validAliasID(aliasID) {
		return RetirementKey{}, fmt.Errorf("V1 alias retirement identity is invalid")
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte(retirementKeyDomain))
	_, _ = digest.Write([]byte(manifestID))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write([]byte(aliasID))
	key := RetirementKey{
		value: retirementKeyPrefix + hex.EncodeToString(digest.Sum(nil)),
		valid: true,
	}
	key.stateSeal = retirementKeySeal(key.value)
	if key.Validate() != nil {
		return RetirementKey{}, fmt.Errorf("V1 alias retirement identity is invalid")
	}
	return key, nil
}

func (key RetirementKey) Validate() error {
	const digestLength = sha256.Size * 2
	if !key.valid || len(key.value) != len(retirementKeyPrefix)+digestLength ||
		key.value[:len(retirementKeyPrefix)] != retirementKeyPrefix ||
		key.stateSeal == [sha256.Size]byte{} ||
		key.stateSeal != retirementKeySeal(key.value) {
		return fmt.Errorf("V1 alias retirement identity is invalid")
	}
	if _, err := hex.DecodeString(key.value[len(retirementKeyPrefix):]); err != nil {
		return fmt.Errorf("V1 alias retirement identity is invalid")
	}
	return nil
}

func (key RetirementKey) Value() (string, error) {
	if key.Validate() != nil {
		return "", fmt.Errorf("V1 alias retirement identity is invalid")
	}
	return key.value, nil
}

func retirementKeySeal(value string) [sha256.Size]byte {
	return sha256.Sum256([]byte(retirementSealDomain + value))
}

func (RetirementKey) String() string   { return v1Redaction }
func (RetirementKey) GoString() string { return v1Redaction }
func (key RetirementKey) Format(state fmt.State, _ rune) {
	writeV1Redaction(state)
}
func (RetirementKey) MarshalJSON() ([]byte, error) { return rejectV1JSON() }
