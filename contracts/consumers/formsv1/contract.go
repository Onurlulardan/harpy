package formsv1

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
)

//go:embed forms-record-list.v1.json
var document []byte

type Contract struct {
	Version         int      `json:"version"`
	ContractID      string   `json:"contract_id"`
	Consumers       []string `json:"consumers"`
	Routes          []Route  `json:"routes"`
	QueryKeys       []string `json:"query_keys"`
	SuccessPointers []string `json:"success_pointers"`
}

type Route struct {
	ID            string          `json:"id"`
	Method        string          `json:"method"`
	Path          string          `json:"path"`
	OperationID   string          `json:"operation_id"`
	SuccessStatus int             `json:"success_status"`
	Errors        []ErrorContract `json:"errors"`
}

type ErrorContract struct {
	Scenario string `json:"scenario"`
	Status   int    `json:"status"`
	Code     string `json:"code"`
}

func Load() (Contract, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var contract Contract
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, fmt.Errorf("decode Forms V1 consumer contract: %w", err)
	}
	if err := rejectTrailingJSON(decoder); err != nil {
		return Contract{}, err
	}
	if err := contract.Validate(); err != nil {
		return Contract{}, err
	}
	return contract.clone(), nil
}

func rejectTrailingJSON(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("Forms V1 consumer contract has trailing JSON")
		}
		return fmt.Errorf("decode Forms V1 consumer contract tail: %w", err)
	}
	return nil
}

func (contract Contract) clone() Contract {
	result := contract
	result.Consumers = append([]string{}, contract.Consumers...)
	result.QueryKeys = append([]string{}, contract.QueryKeys...)
	result.SuccessPointers = append([]string{}, contract.SuccessPointers...)
	result.Routes = append([]Route{}, contract.Routes...)
	for index := range result.Routes {
		result.Routes[index].Errors = append(
			[]ErrorContract{}, contract.Routes[index].Errors...,
		)
	}
	return result
}
