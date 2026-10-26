package resolved

import (
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

type mandatoryValueDocument struct {
	Kind  string `json:"kind"`
	Value any    `json:"value"`
}

func canonicalMandatoryValue(value Value) (json.RawMessage, error) {
	if value.Validate() != nil {
		return nil, fmt.Errorf("resolved REST Query mandatory value is invalid")
	}
	document := mandatoryValueDocument{Kind: value.kind.String()}
	switch value.state {
	case valueStateScalar:
		document.Value = value.scalar
	case valueStateBoolean:
		document.Value = value.boolean
	case valueStateJSON:
		document.Value = json.RawMessage(value.json)
	default:
		return nil, fmt.Errorf("resolved REST Query mandatory value is invalid")
	}
	return attestation.MarshalCanonical(document)
}
