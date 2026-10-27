package api

import "github.com/google/uuid"

func (catalog FieldCatalog) Reference() (CatalogReference, error) {
	if err := catalog.Validate(); err != nil {
		return CatalogReference{}, err
	}
	return catalog.reference, nil
}

func (catalog FieldCatalog) Storage() (ResourceStorage, error) {
	if err := catalog.Validate(); err != nil {
		return ResourceStorage{}, err
	}
	return catalog.storage, nil
}

func (catalog FieldCatalog) Capabilities() (CatalogCapabilities, error) {
	if err := catalog.Validate(); err != nil {
		return CatalogCapabilities{}, err
	}
	return catalog.capabilities, nil
}

func (catalog FieldCatalog) Fields() ([]FieldDefinition, error) {
	if err := catalog.Validate(); err != nil {
		return nil, err
	}
	return cloneFieldDefinitions(catalog.fields), nil
}

func (catalog FieldCatalog) Field(
	id uuid.UUID,
) (FieldDefinition, bool, error) {
	if err := catalog.Validate(); err != nil {
		return FieldDefinition{}, false, err
	}
	for _, field := range catalog.fields {
		if field.id == id {
			return field.clone(), true, nil
		}
	}
	return FieldDefinition{}, false, nil
}

func (catalog FieldCatalog) FieldByPublicName(
	name string,
) (FieldDefinition, bool, error) {
	if err := catalog.Validate(); err != nil || !validFieldPublicName(name) {
		return FieldDefinition{}, false, errInvalidFieldCatalog
	}
	for _, field := range catalog.fields {
		if field.publicName == name {
			return field.clone(), true, nil
		}
	}
	return FieldDefinition{}, false, nil
}

func (catalog FieldCatalog) Relations() ([]RelationDefinition, error) {
	if err := catalog.Validate(); err != nil {
		return nil, err
	}
	return append([]RelationDefinition{}, catalog.relations...), nil
}

func (catalog FieldCatalog) Relation(
	id uuid.UUID,
) (RelationDefinition, bool, error) {
	if err := catalog.Validate(); err != nil {
		return RelationDefinition{}, false, err
	}
	for _, relation := range catalog.relations {
		if relation.relationID == id {
			return relation, true, nil
		}
	}
	return RelationDefinition{}, false, nil
}

func (catalog FieldCatalog) RelationByPublicName(
	name string,
) (RelationDefinition, bool, error) {
	if err := catalog.Validate(); err != nil || !validRelationPublicName(name) {
		return RelationDefinition{}, false, errInvalidFieldCatalog
	}
	for _, relation := range catalog.relations {
		if relation.publicName == name {
			return relation, true, nil
		}
	}
	return RelationDefinition{}, false, nil
}

func (catalog FieldCatalog) DefaultSelect() ([]uuid.UUID, error) {
	if err := catalog.Validate(); err != nil {
		return nil, err
	}
	return append([]uuid.UUID{}, catalog.defaultSelect...), nil
}

func (catalog FieldCatalog) StableID() (uuid.UUID, error) {
	if err := catalog.Validate(); err != nil {
		return uuid.Nil, err
	}
	return catalog.stableID, nil
}

func (catalog FieldCatalog) SearchFields() ([]uuid.UUID, error) {
	if err := catalog.Validate(); err != nil {
		return nil, err
	}
	return append([]uuid.UUID{}, catalog.searchFields...), nil
}
