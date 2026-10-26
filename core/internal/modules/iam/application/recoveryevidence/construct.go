package recoveryevidence

import "fmt"

type Service struct{ store Store }

func New(store Store) (*Service, error) {
	if store == nil {
		return nil, fmt.Errorf("IAM recovery evidence service is unavailable")
	}
	return &Service{store: store}, nil
}
