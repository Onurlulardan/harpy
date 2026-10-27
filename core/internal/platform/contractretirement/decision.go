package contractretirement

import (
	"crypto/sha256"
	"fmt"
	"slices"
)

const decisionSealDomain = "harpy/contract-retirement/decision-seal/v1\x00"

type Decision struct {
	submitted []Key
	retired   []Key
	batchSeal [sha256.Size]byte
	seal      [sha256.Size]byte
}

func SealDecision(batch Batch, retired []Key) (Decision, error) {
	if batch.Validate() != nil {
		return Decision{}, ErrInvalid
	}
	owned := append([]Key{}, retired...)
	slices.SortFunc(owned, compareKeys)
	for index, key := range owned {
		if key.Validate() != nil || index > 0 && compareKeys(owned[index-1], key) >= 0 ||
			!containsKey(batch.keys, key) {
			return Decision{}, ErrInvalid
		}
	}
	decision := Decision{
		submitted: append([]Key{}, batch.keys...), retired: owned, batchSeal: batch.seal,
	}
	decision.seal = sealDecision(decision)
	return decision, nil
}

func (decision Decision) Validate() error {
	if len(decision.submitted) < 1 || decision.batchSeal == [sha256.Size]byte{} ||
		decision.seal != sealDecision(decision) {
		return ErrInvalid
	}
	for index, key := range decision.submitted {
		if key.Validate() != nil || index > 0 && compareKeys(decision.submitted[index-1], key) >= 0 {
			return ErrInvalid
		}
	}
	for index, key := range decision.retired {
		if key.Validate() != nil || index > 0 && compareKeys(decision.retired[index-1], key) >= 0 ||
			!containsKey(decision.submitted, key) {
			return ErrInvalid
		}
	}
	return nil
}

func (decision Decision) Retires(key Key) (bool, error) {
	if decision.Validate() != nil || key.Validate() != nil ||
		!containsKey(decision.submitted, key) {
		return false, ErrInvalid
	}
	return containsKey(decision.retired, key), nil
}

func (decision Decision) RetiredCount() (int, error) {
	if err := decision.Validate(); err != nil {
		return 0, err
	}
	return len(decision.retired), nil
}

func containsKey(keys []Key, target Key) bool {
	_, found := slices.BinarySearchFunc(keys, target, compareKeys)
	return found
}

func sealDecision(decision Decision) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(decisionSealDomain))
	_, _ = hash.Write(decision.batchSeal[:])
	for _, key := range decision.retired {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(key.value))
	}
	var seal [sha256.Size]byte
	copy(seal[:], hash.Sum(nil))
	return seal
}

func (Decision) String() string                          { return redaction }
func (Decision) GoString() string                        { return redaction }
func (decision Decision) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Decision) MarshalJSON() ([]byte, error)            { return rejectJSON() }
