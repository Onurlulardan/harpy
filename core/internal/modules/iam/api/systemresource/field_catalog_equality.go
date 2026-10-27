package systemresource

import "crypto/subtle"

func (catalog FieldCatalog) ExactEqual(other FieldCatalog) (bool, error) {
	if catalog.Validate() != nil || other.Validate() != nil {
		return false, ErrContract
	}
	presentationEqual, err := catalog.presentation.ExactEqual(other.presentation)
	if err != nil {
		return false, ErrContract
	}
	return catalog.resource == other.resource && catalog.resourceID == other.resourceID &&
		catalog.publishedVersionID == other.publishedVersionID && presentationEqual &&
		descriptorsExactEqual(catalog.descriptors, other.descriptors) &&
		subtle.ConstantTimeCompare(catalog.querySHA256[:], other.querySHA256[:]) == 1, nil
}

func descriptorsExactEqual(left, right []FieldDescriptor) bool {
	if len(left) != len(right) {
		return false
	}
	for _, descriptor := range left {
		id, idErr := descriptor.ID()
		name, nameErr := descriptor.PublicName()
		sortable, sortableErr := descriptor.Sortable()
		searchable, searchableErr := descriptor.Searchable()
		required, requiredErr := descriptor.RequiredForCreate()
		maximumBytes, maximumErr := descriptor.MaximumInputBytes()
		other, found := descriptorByID(right, id)
		otherName, otherNameErr := other.PublicName()
		otherSortable, otherSortableErr := other.Sortable()
		otherSearchable, otherSearchableErr := other.Searchable()
		otherRequired, otherRequiredErr := other.RequiredForCreate()
		otherMaximumBytes, otherMaximumErr := other.MaximumInputBytes()
		if idErr != nil || nameErr != nil || sortableErr != nil ||
			searchableErr != nil || requiredErr != nil || !found || otherNameErr != nil ||
			otherSortableErr != nil || otherSearchableErr != nil ||
			otherRequiredErr != nil ||
			maximumErr != nil || otherMaximumErr != nil ||
			name != otherName || sortable != otherSortable ||
			searchable != otherSearchable || required != otherRequired ||
			maximumBytes != otherMaximumBytes {
			return false
		}
	}
	return true
}

func descriptorByID(values []FieldDescriptor, wanted ID) (FieldDescriptor, bool) {
	for _, descriptor := range values {
		id, err := descriptor.ID()
		if err == nil && id == wanted {
			return descriptor, true
		}
	}
	return FieldDescriptor{}, false
}
