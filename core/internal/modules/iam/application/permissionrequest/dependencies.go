package permissionrequest

type Service struct {
	store   Store
	effects DecisionEffects
}

func NewService(
	store Store,
	effects DecisionEffects,
) (*Service, error) {
	if store == nil || effects == nil {
		return nil, ErrUnavailable
	}
	return &Service{store: store, effects: effects}, nil
}
