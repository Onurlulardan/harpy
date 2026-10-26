package main

import (
	"encoding/json"
	"io"
)

type commandResult struct {
	ContractVersion string `json:"contract_version"`
	Action          string `json:"action"`
	Status          string `json:"status"`
	OperationID     string `json:"operation_id"`
	EvidenceSHA256  string `json:"evidence_sha256"`
	Replayed        bool   `json:"replayed"`
}

func writeResult(output io.Writer, result commandResult) error {
	result.ContractVersion = "external-cutover-authority-result.v1"
	result.Status = "succeeded"
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}
