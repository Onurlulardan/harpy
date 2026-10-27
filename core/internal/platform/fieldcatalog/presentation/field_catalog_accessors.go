package presentation

import (
	"crypto/subtle"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
)

func (field Field) ID() (string, error) {
	if field.Validate() != nil {
		return "", ErrInvalid
	}
	return field.id, nil
}

func (field Field) TypeID() (fieldregistry.TypeID, error) {
	if field.Validate() != nil {
		return "", ErrInvalid
	}
	return field.typeID, nil
}

func (field Field) ValueProfile() (fieldregistry.ValueProfile, error) {
	if field.Validate() != nil {
		return "", ErrInvalid
	}
	return field.valueProfile, nil
}

func (field Field) Mode() (FieldMode, error) {
	if field.Validate() != nil {
		return "", ErrInvalid
	}
	return field.mode, nil
}

func (catalog FieldCatalog) Fields() ([]Field, error) {
	if catalog.Validate() != nil {
		return nil, ErrInvalid
	}
	return append([]Field(nil), catalog.fields...), nil
}

func (catalog FieldCatalog) Field(id string) (Field, bool, error) {
	if catalog.Validate() != nil || !stableUUIDText(id) {
		return Field{}, false, ErrInvalid
	}
	for _, field := range catalog.fields {
		if field.id == id {
			return field, true, nil
		}
	}
	return Field{}, false, nil
}

func (catalog FieldCatalog) CanonicalJSON() ([]byte, error) {
	if catalog.Validate() != nil {
		return nil, ErrInvalid
	}
	return append([]byte(nil), catalog.canonicalJSON...), nil
}

func (catalog FieldCatalog) CanonicalSHA256() ([32]byte, error) {
	if catalog.Validate() != nil {
		return [32]byte{}, ErrInvalid
	}
	return catalog.digest, nil
}

func (catalog FieldCatalog) ExactEqual(other FieldCatalog) (bool, error) {
	if catalog.Validate() != nil || other.Validate() != nil {
		return false, ErrInvalid
	}
	return subtle.ConstantTimeCompare(catalog.digest[:], other.digest[:]) == 1 &&
		subtle.ConstantTimeCompare(catalog.canonicalJSON, other.canonicalJSON) == 1, nil
}
