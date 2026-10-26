package v1query

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
)

// PredecessorOrdinal is minted only from a validated page>1 PagePlan.
type PredecessorOrdinal struct {
	page      int64
	limit     int
	index     int64
	stateSeal [sha256.Size]byte
	validated bool
}

func (plan PagePlan) PredecessorOrdinal() (PredecessorOrdinal, bool, error) {
	index, present, err := plan.PredecessorIndex()
	if err != nil || !present {
		return PredecessorOrdinal{}, present, err
	}
	ordinal := PredecessorOrdinal{
		page: plan.page, limit: plan.limit, index: index, validated: true,
	}
	ordinal.stateSeal = sealPredecessorOrdinal(ordinal)
	if ordinal.Validate() != nil {
		return PredecessorOrdinal{}, false, fmt.Errorf("V1 predecessor ordinal is invalid")
	}
	return ordinal, true, nil
}

func (ordinal PredecessorOrdinal) Validate() error {
	if !ordinal.validated || ordinal.page < 2 || ordinal.limit < 1 ||
		ordinal.limit > 500 || ordinal.index < 0 ||
		ordinal.page-1 > math.MaxInt64/int64(ordinal.limit) ||
		ordinal.index != (ordinal.page-1)*int64(ordinal.limit)-1 ||
		ordinal.stateSeal == [sha256.Size]byte{} ||
		ordinal.stateSeal != sealPredecessorOrdinal(ordinal) {
		return fmt.Errorf("V1 predecessor ordinal is invalid")
	}
	return nil
}

func (ordinal PredecessorOrdinal) Limit() (int, error) {
	if ordinal.Validate() != nil {
		return 0, fmt.Errorf("V1 predecessor ordinal is invalid")
	}
	return ordinal.limit, nil
}

func (ordinal PredecessorOrdinal) Index() (int64, error) {
	if ordinal.Validate() != nil {
		return 0, fmt.Errorf("V1 predecessor ordinal is invalid")
	}
	return ordinal.index, nil
}

func sealPredecessorOrdinal(ordinal PredecessorOrdinal) [sha256.Size]byte {
	document := make([]byte, 0, 64)
	document = append(document, "forgecrud.restquery.v1.predecessor-ordinal.v1"...)
	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, uint64(ordinal.page))
	document = append(document, value...)
	binary.BigEndian.PutUint64(value, uint64(ordinal.limit))
	document = append(document, value...)
	binary.BigEndian.PutUint64(value, uint64(ordinal.index))
	document = append(document, value...)
	return sha256.Sum256(document)
}

func (PredecessorOrdinal) String() string   { return v1Redaction }
func (PredecessorOrdinal) GoString() string { return v1Redaction }
func (ordinal PredecessorOrdinal) Format(state fmt.State, _ rune) {
	writeV1Redaction(state)
}
func (PredecessorOrdinal) MarshalJSON() ([]byte, error) { return rejectV1JSON() }
