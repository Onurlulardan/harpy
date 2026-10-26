package systemresource

type Record struct {
	resource  Resource
	id        ID
	revision  int64
	values    FieldValues
	validated bool
}

func NewRecord(
	resource Resource,
	id ID,
	revision int64,
	values FieldValues,
) (Record, error) {
	record := Record{
		resource: resource, id: id, revision: revision,
		values: values, validated: true,
	}
	if record.Validate() != nil {
		return Record{}, ErrContract
	}
	return record, nil
}

func (record Record) Validate() error {
	if !record.validated || !record.resource.Valid() || !record.id.Valid() ||
		record.revision < 1 || record.values.Validate() != nil ||
		record.values.Len() == 0 {
		return ErrContract
	}
	return nil
}

func (record Record) Resource() (Resource, error) {
	if record.Validate() != nil {
		return "", ErrContract
	}
	return record.resource, nil
}

func (record Record) Identity() (ID, int64, error) {
	if record.Validate() != nil {
		return ID{}, 0, ErrContract
	}
	return record.id, record.revision, nil
}

func (record Record) Values() (FieldValues, error) {
	if record.Validate() != nil {
		return FieldValues{}, ErrContract
	}
	return record.values, nil
}
