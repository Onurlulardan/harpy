package contact

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

type Store struct{ driver ports.ContactDriver }

func NewStore(driver ports.ContactDriver) (*Store, error) {
	if !ports.Available(driver) {
		return nil, fmt.Errorf("support contact postgres driver is unavailable")
	}
	return &Store{driver: driver}, nil
}

func (store *Store) ready() bool {
	return store != nil && ports.Available(store.driver)
}
