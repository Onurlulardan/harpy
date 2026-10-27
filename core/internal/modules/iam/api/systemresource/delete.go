package systemresource

type DeleteInput struct {
	resource         Resource
	id               ID
	expectedRevision int64
	idempotencyKey   string
	validated        bool
}

func NewDeleteInput(
	resource Resource,
	id ID,
	expectedRevision int64,
	idempotencyKey string,
) (DeleteInput, error) {
	input := DeleteInput{
		resource: resource, id: id, expectedRevision: expectedRevision,
		idempotencyKey: idempotencyKey, validated: true,
	}
	if input.Validate() != nil {
		return DeleteInput{}, ErrContract
	}
	return input, nil
}

func (input DeleteInput) Validate() error {
	if !input.validated || !input.resource.Valid() || !input.id.Valid() ||
		input.expectedRevision < 1 || !validIdempotencyKey(input.idempotencyKey) {
		return ErrContract
	}
	return nil
}

func (input DeleteInput) Resource() (Resource, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.resource, nil
}

func (input DeleteInput) Identity() (ID, int64, error) {
	if input.Validate() != nil {
		return ID{}, 0, ErrContract
	}
	return input.id, input.expectedRevision, nil
}

func (input DeleteInput) IdempotencyKey() (string, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.idempotencyKey, nil
}
