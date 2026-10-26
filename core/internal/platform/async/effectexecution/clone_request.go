package effectexecution

import (
	"fmt"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
)

func cloneRequest(request effectsv1.Request) (effectsv1.Request, error) {
	document, err := request.MarshalJSON()
	if err != nil {
		return effectsv1.Request{}, fmt.Errorf("encode effect request: %w", err)
	}
	copy, err := effectsv1.DecodeRequest(document)
	if err != nil {
		return effectsv1.Request{}, fmt.Errorf("decode effect request copy: %w", err)
	}
	return copy, nil
}
