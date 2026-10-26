package billingv1adoptionapp

import (
	"bytes"
	"encoding/json"
)

func decodeObject(document []byte, allowed map[string]struct{}) (map[string]json.RawMessage, error) {
	if len(document) == 0 {
		document = []byte("{}")
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	var values map[string]json.RawMessage
	if err := decoder.Decode(&values); err != nil || values == nil || decoder.More() {
		return nil, ErrPlan
	}
	var trailing any
	if decoder.Decode(&trailing) == nil {
		return nil, ErrPlan
	}
	for key := range values {
		if _, ok := allowed[key]; !ok {
			return nil, ErrPlan
		}
	}
	return values, nil
}
