package recovery

type Service struct {
	store    Store
	identity Identity
}

func New(store Store, identity Identity) (*Service, error) {
	if store == nil || identity.EnvironmentID == [16]byte{} ||
		identity.TenantID == [16]byte{} || identity.TenantSetSHA256 == [32]byte{} {
		return nil, ErrUnavailable
	}
	return &Service{store: store, identity: identity}, nil
}
