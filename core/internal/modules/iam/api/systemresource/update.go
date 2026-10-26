package systemresource

type UpdateInput struct {
	resource         Resource
	id               ID
	expectedRevision int64
	values           FieldValues
	idempotencyKey   string
	validated        bool
}

func NewUpdateInput(
	resource Resource,
	id ID,
	expectedRevision int64,
	values FieldValues,
	idempotencyKey string,
) (UpdateInput, error) {
	input := UpdateInput{
		resource: resource, id: id, expectedRevision: expectedRevision,
		values: values, idempotencyKey: idempotencyKey, validated: true,
	}
	if input.Validate() != nil {
		return UpdateInput{}, ErrContract
	}
	return input, nil
}

func (input UpdateInput) Validate() error {
	if !input.validated || !input.resource.Valid() || !input.id.Valid() ||
		input.expectedRevision < 1 || input.values.Validate() != nil ||
		input.values.Len() == 0 || !validIdempotencyKey(input.idempotencyKey) {
		return ErrContract
	}
	return nil
}

func (input UpdateInput) Resource() (Resource, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.resource, nil
}

func (input UpdateInput) Identity() (ID, int64, error) {
	if input.Validate() != nil {
		return ID{}, 0, ErrContract
	}
	return input.id, input.expectedRevision, nil
}

func (input UpdateInput) Values() (FieldValues, error) {
	if input.Validate() != nil {
		return FieldValues{}, ErrContract
	}
	return input.values, nil
}

func (input UpdateInput) IdempotencyKey() (string, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.idempotencyKey, nil
}
