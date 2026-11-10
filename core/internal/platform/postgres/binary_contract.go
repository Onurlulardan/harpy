package postgres

import (
	"encoding/json"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func loadEmbeddedBinaryContract() (runtimev1.BinarySchemaContract, error) {
	document, err := runtimev1.ReadEmbeddedBinarySchemaContract()
	if err != nil {
		return runtimev1.BinarySchemaContract{}, err
	}
	var contract runtimev1.BinarySchemaContract
	if err = json.Unmarshal(document, &contract); err != nil {
		return runtimev1.BinarySchemaContract{}, fmt.Errorf("decode binary schema contract: %w", err)
	}
	if err = contract.Validate(); err != nil {
		return runtimev1.BinarySchemaContract{}, err
	}
	return contract, nil
}
