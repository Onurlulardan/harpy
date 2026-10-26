package systemresource

type GetInput struct {
	resource  Resource
	id        ID
	validated bool
}

func NewGetInput(resource Resource, id ID) (GetInput, error) {
	input := GetInput{resource: resource, id: id, validated: true}
	if input.Validate() != nil {
		return GetInput{}, ErrContract
	}
	return input, nil
}

func (input GetInput) Validate() error {
	if !input.validated || !input.resource.Valid() || !input.id.Valid() {
		return ErrContract
	}
	return nil
}

func (input GetInput) Resource() (Resource, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.resource, nil
}

func (input GetInput) ID() (ID, error) {
	if input.Validate() != nil {
		return ID{}, ErrContract
	}
	return input.id, nil
}
