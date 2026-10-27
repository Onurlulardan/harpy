package systemresource

import formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"

type FieldCatalog struct {
	resource           Resource
	resourceID         ID
	publishedVersionID ID
	querySHA256        [32]byte
	presentation       formpresentation.FieldCatalog
	descriptors        []FieldDescriptor
	validated          bool
}

func NewFieldCatalog(
	resource Resource,
	resourceID ID,
	publishedVersionID ID,
	querySHA256 [32]byte,
	presentation formpresentation.FieldCatalog,
	descriptors []FieldDescriptor,
) (FieldCatalog, error) {
	catalog := FieldCatalog{
		resource: resource, resourceID: resourceID,
		publishedVersionID: publishedVersionID, querySHA256: querySHA256,
		presentation: presentation,
		descriptors:  append([]FieldDescriptor(nil), descriptors...), validated: true,
	}
	if catalog.Validate() != nil {
		return FieldCatalog{}, ErrContract
	}
	return catalog, nil
}

func (catalog FieldCatalog) Validate() error {
	wantedResourceID, idErr := catalog.resource.CatalogResourceID()
	fields, fieldsErr := catalog.presentation.Fields()
	if !catalog.validated || idErr != nil || !catalog.resourceID.Valid() ||
		catalog.resourceID.String() != wantedResourceID ||
		!catalog.publishedVersionID.Valid() || catalog.querySHA256 == [32]byte{} ||
		catalog.presentation.Validate() != nil || fieldsErr != nil || len(fields) == 0 ||
		!descriptorsMatchFields(catalog.descriptors, fields) {
		return ErrContract
	}
	return nil
}

func descriptorsMatchFields(
	descriptors []FieldDescriptor,
	fields []formpresentation.Field,
) bool {
	if len(descriptors) != len(fields) {
		return false
	}
	seenIDs := make(map[string]struct{}, len(descriptors))
	seenNames := make(map[string]struct{}, len(descriptors))
	for _, descriptor := range descriptors {
		id, idErr := descriptor.ID()
		name, nameErr := descriptor.PublicName()
		if idErr != nil || nameErr != nil {
			return false
		}
		if _, duplicate := seenIDs[id.String()]; duplicate {
			return false
		}
		if _, duplicate := seenNames[name]; duplicate {
			return false
		}
		if _, found, err := fieldByID(fields, id.String()); err != nil || !found {
			return false
		}
		seenIDs[id.String()] = struct{}{}
		seenNames[name] = struct{}{}
	}
	return true
}

func fieldByID(
	fields []formpresentation.Field,
	id string,
) (formpresentation.Field, bool, error) {
	for _, field := range fields {
		fieldID, err := field.ID()
		if err != nil {
			return formpresentation.Field{}, false, err
		}
		if fieldID == id {
			return field, true, nil
		}
	}
	return formpresentation.Field{}, false, nil
}

func (catalog FieldCatalog) Resource() (Resource, error) {
	if catalog.Validate() != nil {
		return "", ErrContract
	}
	return catalog.resource, nil
}

func (catalog FieldCatalog) Reference() (ID, ID, [32]byte, error) {
	if catalog.Validate() != nil {
		return ID{}, ID{}, [32]byte{}, ErrContract
	}
	return catalog.resourceID, catalog.publishedVersionID, catalog.querySHA256, nil
}

func (catalog FieldCatalog) Presentation() (formpresentation.FieldCatalog, error) {
	if catalog.Validate() != nil {
		return formpresentation.FieldCatalog{}, ErrContract
	}
	return catalog.presentation, nil
}

func (catalog FieldCatalog) Descriptors() ([]FieldDescriptor, error) {
	if catalog.Validate() != nil {
		return nil, ErrContract
	}
	return append([]FieldDescriptor(nil), catalog.descriptors...), nil
}

func (catalog FieldCatalog) Descriptor(fieldID ID) (FieldDescriptor, bool, error) {
	if catalog.Validate() != nil || !fieldID.Valid() {
		return FieldDescriptor{}, false, ErrContract
	}
	for _, descriptor := range catalog.descriptors {
		id, err := descriptor.ID()
		if err != nil {
			return FieldDescriptor{}, false, ErrContract
		}
		if id == fieldID {
			return descriptor, true, nil
		}
	}
	return FieldDescriptor{}, false, nil
}

func (catalog FieldCatalog) Editable(fieldID ID) (bool, error) {
	if catalog.Validate() != nil || !fieldID.Valid() {
		return false, ErrContract
	}
	field, found, err := catalog.presentation.Field(fieldID.String())
	if err != nil || !found {
		return false, ErrContract
	}
	mode, err := field.Mode()
	if err != nil {
		return false, ErrContract
	}
	return mode == formpresentation.FieldEditable, nil
}
