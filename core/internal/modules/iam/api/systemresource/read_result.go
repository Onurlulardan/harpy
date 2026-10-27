package systemresource

type ReadResult struct {
	record    Record
	catalog   FieldCatalog
	relations RelationProjection
	validated bool
}

func NewReadResult(record Record, catalog FieldCatalog) (ReadResult, error) {
	resource, err := record.Resource()
	if err != nil {
		return ReadResult{}, ErrContract
	}
	relations, err := EmptyRelationProjection(resource)
	if err != nil {
		return ReadResult{}, ErrContract
	}
	return NewReadResultWithRelations(record, catalog, relations)
}

func NewReadResultWithRelations(
	record Record,
	catalog FieldCatalog,
	relations RelationProjection,
) (ReadResult, error) {
	result := ReadResult{
		record: record, catalog: catalog,
		relations: relations, validated: true,
	}
	if result.Validate() != nil {
		return ReadResult{}, ErrContract
	}
	return result, nil
}

func (result ReadResult) Validate() error {
	recordResource, recordErr := result.record.Resource()
	catalogResource, catalogErr := result.catalog.Resource()
	relationResource, _, relationErr := result.relations.Values()
	if !result.validated || recordErr != nil || catalogErr != nil ||
		relationErr != nil || recordResource != catalogResource ||
		recordResource != relationResource {
		return ErrContract
	}
	return nil
}

func (result ReadResult) Record() (Record, error) {
	if result.Validate() != nil {
		return Record{}, ErrContract
	}
	return result.record, nil
}

func (result ReadResult) Catalog() (FieldCatalog, error) {
	if result.Validate() != nil {
		return FieldCatalog{}, ErrContract
	}
	return result.catalog, nil
}

func (result ReadResult) Relations() (RelationProjection, error) {
	if result.Validate() != nil {
		return RelationProjection{}, ErrContract
	}
	return result.relations, nil
}
