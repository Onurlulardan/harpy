package platformlicense

import "fmt"

type Service struct{ store Store }

func NewService(store Store) (*Service, error) {
	if store == nil {
		return nil, fmt.Errorf("billing platform license store is required")
	}
	return &Service{store: store}, nil
}
