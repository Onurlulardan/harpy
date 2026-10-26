package api

import (
	"errors"
	"fmt"
)

var errInvalidCatalogCapabilities = errors.New("REST Query catalog capabilities are invalid")

type CatalogCapabilities struct {
	exactCountEnabled bool
	validated         bool
}

func NewCatalogCapabilities(exactCountEnabled bool) (CatalogCapabilities, error) {
	capabilities := CatalogCapabilities{
		exactCountEnabled: exactCountEnabled, validated: true,
	}
	if err := capabilities.Validate(); err != nil {
		return CatalogCapabilities{}, err
	}
	return capabilities, nil
}

func (capabilities CatalogCapabilities) Validate() error {
	if !capabilities.validated {
		return errInvalidCatalogCapabilities
	}
	return nil
}

func (capabilities CatalogCapabilities) ExactCountEnabled() (bool, error) {
	if err := capabilities.Validate(); err != nil {
		return false, err
	}
	return capabilities.exactCountEnabled, nil
}

func (CatalogCapabilities) String() string   { return queryValueRedaction }
func (CatalogCapabilities) GoString() string { return queryValueRedaction }
func (CatalogCapabilities) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (CatalogCapabilities) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
