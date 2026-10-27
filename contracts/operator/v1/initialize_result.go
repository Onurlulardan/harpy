package operatorv1

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
)

const InitializeResultContract = "operator-initialize-result.v1"

type InitializeStatus string

const (
	InitializeCompleted InitializeStatus = "completed"
	InitializeNoop      InitializeStatus = "completed_noop"
)

type InitializeResult struct {
	ContractVersion string           `json:"contract_version"`
	Status          InitializeStatus `json:"status"`
	OperationID     string           `json:"operation_id"`
	TenantID        string           `json:"tenant_id"`
}

func NewInitializeResult(
	status InitializeStatus,
	operationID uuid.UUID,
	tenantID uuid.UUID,
) (InitializeResult, error) {
	result := InitializeResult{ContractVersion: InitializeResultContract,
		Status: status, OperationID: operationID.String(), TenantID: tenantID.String()}
	if err := result.Validate(); err != nil {
		return InitializeResult{}, err
	}
	return result, nil
}

func (result InitializeResult) Validate() error {
	operationID, operationErr := uuid.Parse(result.OperationID)
	tenantID, tenantErr := uuid.Parse(result.TenantID)
	if result.ContractVersion != InitializeResultContract ||
		result.Status != InitializeCompleted && result.Status != InitializeNoop ||
		operationErr != nil || tenantErr != nil || operationID == uuid.Nil || tenantID == uuid.Nil {
		return fmt.Errorf("operator initialize result is invalid")
	}
	return nil
}

func WriteInitializeResult(writer io.Writer, result InitializeResult) error {
	if writer == nil || result.Validate() != nil {
		return fmt.Errorf("operator initialize result is invalid")
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}
