package systemresource

const maximumPasswordBytes = 1024

type CreateInput struct {
	resource       Resource
	values         FieldValues
	password       []byte
	idempotencyKey string
	validated      bool
}

func NewCreateInput(
	resource Resource,
	values FieldValues,
	password []byte,
	idempotencyKey string,
) (CreateInput, error) {
	input := CreateInput{
		resource: resource, values: values, password: append([]byte(nil), password...),
		idempotencyKey: idempotencyKey, validated: true,
	}
	if input.Validate() != nil {
		input.Destroy()
		return CreateInput{}, ErrContract
	}
	return input, nil
}

func (input CreateInput) Validate() error {
	passwordValid := input.resource == Users && len(input.password) > 0 &&
		len(input.password) <= maximumPasswordBytes ||
		input.resource != Users && len(input.password) == 0
	if !input.validated || !input.resource.Valid() || input.values.Validate() != nil ||
		input.values.Len() == 0 || !passwordValid ||
		!validIdempotencyKey(input.idempotencyKey) {
		return ErrContract
	}
	return nil
}

func (input CreateInput) Resource() (Resource, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.resource, nil
}

func (input CreateInput) Values() (FieldValues, error) {
	if input.Validate() != nil {
		return FieldValues{}, ErrContract
	}
	return input.values, nil
}

func (input CreateInput) Password() ([]byte, error) {
	if input.Validate() != nil {
		return nil, ErrContract
	}
	return append([]byte(nil), input.password...), nil
}

func (input CreateInput) IdempotencyKey() (string, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.idempotencyKey, nil
}

func (input *CreateInput) Destroy() {
	if input == nil {
		return
	}
	clear(input.password)
	input.password = nil
	input.validated = false
}
