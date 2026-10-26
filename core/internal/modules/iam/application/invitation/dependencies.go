package invitation

type Service struct {
	store      Store
	normalizer EmailNormalizer
	effects    Effects
}

func NewService(store Store, normalizer EmailNormalizer, effects Effects) (*Service, error) {
	if store == nil || normalizer == nil || effects == nil {
		return nil, ErrUnavailable
	}
	return &Service{store: store, normalizer: normalizer, effects: effects}, nil
}
