package presentation

import (
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type catalogFieldWire struct {
	ID           string    `json:"id"`
	Mode         FieldMode `json:"mode"`
	TypeID       string    `json:"type_id"`
	ValueProfile string    `json:"value_profile"`
}

type fieldCatalogWire struct {
	Contract string             `json:"contract"`
	Fields   []catalogFieldWire `json:"fields"`
}

func encodeFieldCatalog(fields []Field) ([]byte, error) {
	wire := fieldCatalogWire{
		Contract: FieldCatalogContract,
		Fields:   make([]catalogFieldWire, len(fields)),
	}
	for index, field := range fields {
		if field.Validate() != nil {
			return nil, ErrInvalid
		}
		wire.Fields[index] = catalogFieldWire{
			ID: field.id, Mode: field.mode, TypeID: string(field.typeID),
			ValueProfile: string(field.valueProfile),
		}
	}
	document, err := json.Marshal(wire)
	if err != nil {
		return nil, ErrInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 1<<20)
	if err != nil {
		return nil, ErrInvalid
	}
	return canonical, nil
}
