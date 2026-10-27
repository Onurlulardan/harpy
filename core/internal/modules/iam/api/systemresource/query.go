package systemresource

type QueryRequest interface {
	Validate() error
}

type QueryInput struct {
	resource  Resource
	request   QueryRequest
	validated bool
}

func NewQueryInput(
	resource Resource,
	request QueryRequest,
) (QueryInput, error) {
	input := QueryInput{resource: resource, request: request, validated: true}
	if input.Validate() != nil {
		return QueryInput{}, ErrContract
	}
	return input, nil
}

func (input QueryInput) Validate() error {
	if !input.validated || !input.resource.Valid() || input.request == nil ||
		input.request.Validate() != nil {
		return ErrContract
	}
	return nil
}

func (input QueryInput) Resource() (Resource, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.resource, nil
}

func (input QueryInput) Request() (QueryRequest, error) {
	if input.Validate() != nil {
		return nil, ErrContract
	}
	return input.request, nil
}
