package application

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type DeliveryQueue struct {
	store ports.DeliveryStore
}

func NewDeliveryQueue(store ports.DeliveryStore) (*DeliveryQueue, error) {
	if store == nil {
		return nil, fmt.Errorf("platform authority delivery queue is unavailable")
	}
	return &DeliveryQueue{store: store}, nil
}
