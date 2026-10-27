package v1adoption

import (
	"bytes"
	"fmt"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func ValidateCutoverBatch(
	contractID string,
	payload cutoverv1.ChangeBatchPayload,
) error {
	if !applyContract(contractID) && contractID != CompletionContractID {
		return nil
	}
	if payload.CheckpointState != "complete" || len(payload.Changes) != 1 {
		return fmt.Errorf("V1 environment cutover batch must be atomic")
	}
	change := payload.Changes[0]
	if change.Ordinal != 1 || change.Kind != "insert" ||
		!bytes.Equal(bytes.TrimSpace(change.Before), []byte("null")) ||
		len(bytes.TrimSpace(change.After)) == 0 ||
		bytes.Equal(bytes.TrimSpace(change.After), []byte("null")) {
		return fmt.Errorf("V1 environment cutover change is invalid")
	}
	if err := validateOperationKey(payload.OperationID, change.Key); err != nil {
		return err
	}
	if applyContract(contractID) {
		return validateApplyDocument(payload, change.After)
	}
	return validateCompletionDocument(payload, change.After)
}

func validateOperationKey(operationID string, document []byte) error {
	if string(bytes.TrimSpace(document)) != `{"operation_id":"`+operationID+`"}` {
		return fmt.Errorf("V1 environment cutover key differs")
	}
	return nil
}
