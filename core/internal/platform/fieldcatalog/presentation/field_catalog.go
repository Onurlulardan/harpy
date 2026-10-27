package presentation

import (
	"bytes"
	"crypto/sha256"
	"sort"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
)

const maximumCatalogFields = 500

type Field struct {
	id           string
	typeID       fieldregistry.TypeID
	valueProfile fieldregistry.ValueProfile
	mode         FieldMode
	validated    bool
}

type FieldCatalog struct {
	fields        []Field
	canonicalJSON []byte
	digest        [sha256.Size]byte
	validated     bool
}

func NewField(
	id string,
	typeID fieldregistry.TypeID,
	profile fieldregistry.ValueProfile,
	mode FieldMode,
) (Field, error) {
	field := Field{
		id: id, typeID: typeID, valueProfile: profile,
		mode: mode, validated: true,
	}
	if field.Validate() != nil {
		return Field{}, ErrInvalid
	}
	return field, nil
}

func NewFieldCatalog(input []Field) (FieldCatalog, error) {
	if input == nil || len(input) > maximumCatalogFields {
		return FieldCatalog{}, ErrInvalid
	}
	fields := append([]Field(nil), input...)
	sort.Slice(fields, func(left, right int) bool {
		return fields[left].id < fields[right].id
	})
	for index, field := range fields {
		if field.Validate() != nil || index > 0 && fields[index-1].id == field.id {
			return FieldCatalog{}, ErrInvalid
		}
	}
	canonical, err := encodeFieldCatalog(fields)
	if err != nil {
		return FieldCatalog{}, err
	}
	return FieldCatalog{
		fields: fields, canonicalJSON: canonical,
		digest: sha256.Sum256(canonical), validated: true,
	}, nil
}

func (field Field) Validate() error {
	if !field.validated || !stableUUIDText(field.id) ||
		!fieldProfileValid(field.typeID, field.valueProfile) ||
		field.mode != FieldEditable && field.mode != FieldReadOnly {
		return ErrInvalid
	}
	return nil
}

func (catalog FieldCatalog) Validate() error {
	if !catalog.validated || catalog.fields == nil || catalog.canonicalJSON == nil ||
		catalog.digest == [sha256.Size]byte{} {
		return ErrInvalid
	}
	canonical, err := encodeFieldCatalog(catalog.fields)
	if err != nil || !bytes.Equal(canonical, catalog.canonicalJSON) ||
		sha256.Sum256(canonical) != catalog.digest {
		return ErrInvalid
	}
	return nil
}
