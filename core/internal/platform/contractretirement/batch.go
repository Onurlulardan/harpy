package contractretirement

import (
	"crypto/sha256"
	"fmt"
	"slices"
)

const (
	batchLimit      = 256
	batchSealDomain = "harpy/contract-retirement/batch-seal/v1\x00"
)

type Batch struct {
	identity Identity
	keys     []Key
	seal     [sha256.Size]byte
}

func NewBatch(identity Identity, keys []Key) (Batch, error) {
	owned := append([]Key{}, keys...)
	if identity.Validate() != nil || len(owned) < 1 || len(owned) > batchLimit {
		return Batch{}, ErrInvalid
	}
	for _, key := range owned {
		if key.Validate() != nil {
			return Batch{}, ErrInvalid
		}
	}
	slices.SortFunc(owned, compareKeys)
	for index := 1; index < len(owned); index++ {
		if owned[index-1].value == owned[index].value {
			return Batch{}, ErrInvalid
		}
	}
	batch := Batch{identity: identity, keys: owned}
	batch.seal = sealBatch(identity, owned)
	return batch, nil
}

func (batch Batch) Validate() error {
	if batch.identity.Validate() != nil || len(batch.keys) < 1 ||
		len(batch.keys) > batchLimit || batch.seal != sealBatch(batch.identity, batch.keys) {
		return ErrInvalid
	}
	for index, key := range batch.keys {
		if key.Validate() != nil || index > 0 && compareKeys(batch.keys[index-1], key) >= 0 {
			return ErrInvalid
		}
	}
	return nil
}

func (batch Batch) DatabaseValues() (Identity, []string, error) {
	if err := batch.Validate(); err != nil {
		return Identity{}, nil, err
	}
	values := make([]string, len(batch.keys))
	for index := range batch.keys {
		values[index] = batch.keys[index].value
	}
	return batch.identity, values, nil
}

func compareKeys(left, right Key) int { return slices.Compare([]byte(left.value), []byte(right.value)) }

func sealBatch(identity Identity, keys []Key) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(batchSealDomain))
	_, _ = hash.Write(identity.seal[:])
	for _, key := range keys {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(key.value))
	}
	var seal [sha256.Size]byte
	copy(seal[:], hash.Sum(nil))
	return seal
}

func (Batch) String() string                       { return redaction }
func (Batch) GoString() string                     { return redaction }
func (batch Batch) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Batch) MarshalJSON() ([]byte, error)         { return rejectJSON() }
