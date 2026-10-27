package executor

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/internal/jsonscan"
	"github.com/google/uuid"
)

type materialRecord struct {
	id        uuid.UUID
	document  []byte
	order     [][]byte
	stateSeal [sha256.Size]byte
	valid     bool
}

func newMaterialRecord(
	id uuid.UUID,
	document []byte,
	order [][]byte,
) (materialRecord, error) {
	record := materialRecord{
		id: id, document: append([]byte(nil), document...),
		order: cloneMaterialOrder(order), valid: true,
	}
	record.stateSeal = sealMaterialRecord(record)
	if record.validate() != nil {
		return materialRecord{}, ErrContract
	}
	return record, nil
}

func (record materialRecord) validate() error {
	if !record.valid || record.id == uuid.Nil || record.id.Variant() != uuid.RFC4122 ||
		!validMaterialDocument(record.document) ||
		len(record.order) < 1 || len(record.order) > 6 ||
		record.stateSeal == [sha256.Size]byte{} ||
		record.stateSeal != sealMaterialRecord(record) {
		return ErrContract
	}
	for _, value := range record.order {
		if len(value) < 1 || len(value) > 2048 || !json.Valid(value) {
			return ErrContract
		}
	}
	return nil
}

func validMaterialDocument(document []byte) bool {
	if len(document) < 2 || document[0] != '{' || document[len(document)-1] != '}' {
		return false
	}
	scan, err := jsonscan.Scan(document, jsonscan.MaximumNesting)
	if err != nil || len(scan.DuplicatePointers()) != 0 {
		return false
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(document, &object) != nil || object == nil {
		return false
	}
	_, reserved := object["_included"]
	return !reserved
}

func cloneMaterialOrder(values [][]byte) [][]byte {
	result := make([][]byte, len(values))
	for index, value := range values {
		result[index] = append([]byte(nil), value...)
	}
	return result
}
