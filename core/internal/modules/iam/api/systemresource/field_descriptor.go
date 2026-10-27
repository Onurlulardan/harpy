package systemresource

import (
	"strings"
	"unicode/utf8"
)

type FieldDescriptor struct {
	id                ID
	publicName        string
	sortable          bool
	searchable        bool
	createRequired    bool
	maximumInputBytes int
	validated         bool
}

func NewFieldDescriptor(id ID, publicName string) (FieldDescriptor, error) {
	return NewFieldDescriptorWithContract(id, publicName, false, false, false, 0)
}

func NewFieldDescriptorWithCapabilities(
	id ID,
	publicName string,
	sortable bool,
	searchable bool,
) (FieldDescriptor, error) {
	return NewFieldDescriptorWithContract(
		id, publicName, sortable, searchable, false, 0,
	)
}

func NewFieldDescriptorWithContract(
	id ID,
	publicName string,
	sortable bool,
	searchable bool,
	createRequired bool,
	maximumInputBytes int,
) (FieldDescriptor, error) {
	descriptor := FieldDescriptor{
		id: id, publicName: publicName, sortable: sortable,
		searchable: searchable, createRequired: createRequired,
		maximumInputBytes: maximumInputBytes, validated: true,
	}
	if descriptor.Validate() != nil {
		return FieldDescriptor{}, ErrContract
	}
	return descriptor, nil
}

func (descriptor FieldDescriptor) Validate() error {
	if !descriptor.validated || !descriptor.id.Valid() ||
		!validPublicFieldName(descriptor.publicName) ||
		descriptor.maximumInputBytes < 0 ||
		descriptor.maximumInputBytes > 1<<20 {
		return ErrContract
	}
	return nil
}

func (descriptor FieldDescriptor) ID() (ID, error) {
	if descriptor.Validate() != nil {
		return ID{}, ErrContract
	}
	return descriptor.id, nil
}

func (descriptor FieldDescriptor) PublicName() (string, error) {
	if descriptor.Validate() != nil {
		return "", ErrContract
	}
	return descriptor.publicName, nil
}

func (descriptor FieldDescriptor) Sortable() (bool, error) {
	if descriptor.Validate() != nil {
		return false, ErrContract
	}
	return descriptor.sortable, nil
}

func (descriptor FieldDescriptor) Searchable() (bool, error) {
	if descriptor.Validate() != nil {
		return false, ErrContract
	}
	return descriptor.searchable, nil
}

func (descriptor FieldDescriptor) RequiredForCreate() (bool, error) {
	if descriptor.Validate() != nil {
		return false, ErrContract
	}
	return descriptor.createRequired, nil
}

func (descriptor FieldDescriptor) MaximumInputBytes() (int, error) {
	if descriptor.Validate() != nil {
		return 0, ErrContract
	}
	return descriptor.maximumInputBytes, nil
}

func validPublicFieldName(value string) bool {
	return value != "" && value != "_included" && utf8.ValidString(value) &&
		!strings.ContainsRune(value, '\x00')
}
