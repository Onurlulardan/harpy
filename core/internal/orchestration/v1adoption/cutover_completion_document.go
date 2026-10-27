package v1adoption

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/google/uuid"
)

const (
	CompletionContractID      = cutoverv1.PlatformV1EnvironmentCompletionContractID
	CompletionContractVersion = "platform-v1-environment-completion.v1"
)

type CompletionDocument struct {
	ContractVersion        string    `json:"contract_version"`
	OperationID            uuid.UUID `json:"operation_id"`
	EnvironmentID          uuid.UUID `json:"environment_id"`
	PlanSHA256             string    `json:"plan_sha256"`
	FrontendContractSHA256 string    `json:"frontend_contract_sha256"`
	MCPContractSHA256      string    `json:"mcp_contract_sha256"`
	LegacyBooleanHits      int64     `json:"legacy_boolean_hits"`
	ModuleMigrationHits    int64     `json:"module_migration_hits"`
	ZeroHitSince           time.Time `json:"zero_hit_since"`
	CompletedAt            time.Time `json:"completed_at"`
}

func validateCompletionDocument(
	payload cutoverv1.ChangeBatchPayload,
	document []byte,
) error {
	var value CompletionDocument
	if err := decodeStrictJSON(document, &value); err != nil {
		return fmt.Errorf("V1 environment completion document is invalid: %w", err)
	}
	operationID, err := uuid.Parse(payload.OperationID)
	if err != nil || value.ContractVersion != CompletionContractVersion ||
		value.OperationID != operationID || payload.ScopeKind != "global" ||
		value.EnvironmentID.String() != payload.ScopeID {
		return fmt.Errorf("V1 environment completion authority differs")
	}
	if !validSHA(value.PlanSHA256) || !validSHA(value.FrontendContractSHA256) ||
		!validSHA(value.MCPContractSHA256) || value.LegacyBooleanHits != 0 ||
		value.ModuleMigrationHits != 0 || value.ZeroHitSince.Location() != time.UTC ||
		value.CompletedAt.Location() != time.UTC ||
		value.ZeroHitSince.After(value.CompletedAt.Add(-30*24*time.Hour)) {
		return fmt.Errorf("V1 environment completion evidence is invalid")
	}
	return nil
}
