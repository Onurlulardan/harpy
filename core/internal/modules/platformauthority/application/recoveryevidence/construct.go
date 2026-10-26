package recoveryevidence

import "fmt"

type Service struct{ store Store }

func New(store Store) (*Service, error) {
	if store == nil {
		return nil, fmt.Errorf("platform authority recovery evidence service is unavailable")
	}
	return &Service{store: store}, nil
}
