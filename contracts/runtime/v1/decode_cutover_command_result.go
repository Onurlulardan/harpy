package runtimev1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const MaximumCutoverCommandResultBytes = 8192

func DecodeCutoverCommandResult(document []byte) (CutoverCommandResult, error) {
	if len(bytes.TrimSpace(document)) == 0 || len(document) > MaximumCutoverCommandResultBytes {
		return CutoverCommandResult{}, fmt.Errorf("cutover command result size is invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var result CutoverCommandResult
	if err := decoder.Decode(&result); err != nil {
		return CutoverCommandResult{}, fmt.Errorf("decode cutover command result: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return CutoverCommandResult{}, fmt.Errorf("cutover command result has trailing content")
	}
	if err := result.Validate(); err != nil {
		return CutoverCommandResult{}, err
	}
	return result, nil
}
