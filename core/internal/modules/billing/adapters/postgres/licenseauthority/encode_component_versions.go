package licenseauthority

import (
	"encoding/json"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

type componentVersionWire struct {
	Component string `json:"component"`
	Version   string `json:"version"`
}

func encodeComponentVersions(values []licenseapp.ComponentVersion) ([]byte, error) {
	wire := make([]componentVersionWire, len(values))
	for index, value := range values {
		wire[index] = componentVersionWire{
			Component: value.Component, Version: value.Version,
		}
	}
	return json.Marshal(wire)
}
