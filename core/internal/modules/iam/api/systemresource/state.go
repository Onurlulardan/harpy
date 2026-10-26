package systemresource

type State string

const (
	StateInactive  State = "inactive"
	StateRetired   State = "retired"
	StateSuspended State = "suspended"
)

type StateInput struct {
	resource         Resource
	id               ID
	expectedRevision int64
	state            State
	idempotencyKey   string
	validated        bool
}

func NewStateInput(
	resource Resource,
	id ID,
	expectedRevision int64,
	state State,
	idempotencyKey string,
) (StateInput, error) {
	input := StateInput{
		resource: resource, id: id, expectedRevision: expectedRevision,
		state: state, idempotencyKey: idempotencyKey, validated: true,
	}
	if input.Validate() != nil {
		return StateInput{}, ErrContract
	}
	return input, nil
}

func (input StateInput) Validate() error {
	expectedState := map[Resource]State{
		Users: StateInactive, Roles: StateRetired, Organizations: StateSuspended,
	}[input.resource]
	if !input.validated || !input.resource.Valid() || !input.id.Valid() ||
		input.expectedRevision < 1 || input.state != expectedState ||
		!validIdempotencyKey(input.idempotencyKey) {
		return ErrContract
	}
	return nil
}

func (input StateInput) Resource() (Resource, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.resource, nil
}

func (input StateInput) Identity() (ID, int64, error) {
	if input.Validate() != nil {
		return ID{}, 0, ErrContract
	}
	return input.id, input.expectedRevision, nil
}

func (input StateInput) State() (State, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.state, nil
}

func (input StateInput) IdempotencyKey() (string, error) {
	if input.Validate() != nil {
		return "", ErrContract
	}
	return input.idempotencyKey, nil
}
