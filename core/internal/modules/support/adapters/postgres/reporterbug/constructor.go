package reporterbug

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

type Store struct{ driver ports.BugDriver }

func NewStore(driver ports.BugDriver) (*Store, error) {
	if !ports.Available(driver) {
		return nil, fmt.Errorf("support reporter bug postgres driver is unavailable")
	}
	return &Store{driver: driver}, nil
}

func (store *Store) ready() bool { return store != nil && ports.Available(store.driver) }
