package nativequery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func Parse(document []byte) (Plan, bool, error) {
	var header struct {
		Contract string `json:"contract"`
	}
	if json.Unmarshal(document, &header) != nil || header.Contract != ContractV1 {
		return Plan{}, false, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var plan Plan
	if err := decoder.Decode(&plan); err != nil ||
		decoder.Decode(&struct{}{}) != io.EOF || plan.Validate() != nil {
		return Plan{}, true, fmt.Errorf("Analytics native query document is invalid")
	}
	for index := range plan.Sources {
		plan.Sources[index].Query = append(
			json.RawMessage(nil), plan.Sources[index].Query...,
		)
	}
	return plan, true, nil
}

func Encode(plan Plan) ([]byte, error) {
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	document, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("encode Analytics native query: %w", err)
	}
	return document, nil
}
