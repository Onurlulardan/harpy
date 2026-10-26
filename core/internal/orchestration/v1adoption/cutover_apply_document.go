package v1adoption

import (
	"fmt"
	"reflect"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/google/uuid"
)

const (
	ApplyContractID      = cutoverv1.PlatformV1EnvironmentPlanContractID
	ApplyContractVersion = "platform-v1-environment-adoption.v1"
)

type AuthorityIDs struct {
	SubjectID        uuid.UUID `json:"subject_id"`
	OwnerRoleID      uuid.UUID `json:"owner_role_id"`
	RolePermissionID uuid.UUID `json:"role_permission_id"`
	OwnerBindingID   uuid.UUID `json:"owner_binding_id"`
}

func applyContract(contractID string) bool {
	return contractID == cutoverv1.PlatformV1EnvironmentPlanContractID ||
		contractID == cutoverv1.TenancyV1EnvironmentAdoptionContractID ||
		contractID == cutoverv1.IAMV1ProtectedAuthorityAdoptionContractID ||
		contractID == cutoverv1.BillingV1BaseAdoptionContractID
}

type ApplyDocument struct {
	ContractVersion string       `json:"contract_version"`
	OperationID     uuid.UUID    `json:"operation_id"`
	Inventory       Inventory    `json:"inventory"`
	Plan            Plan         `json:"plan"`
	AuthorityIDs    AuthorityIDs `json:"authority_ids"`
}

func validateApplyDocument(
	payload cutoverv1.ChangeBatchPayload,
	document []byte,
) error {
	var value ApplyDocument
	if err := decodeStrictJSON(document, &value); err != nil {
		return fmt.Errorf("V1 environment adoption document is invalid: %w", err)
	}
	operationID, err := uuid.Parse(payload.OperationID)
	if err != nil || value.ContractVersion != ApplyContractVersion ||
		value.OperationID != operationID || payload.ScopeKind != "global" ||
		value.Inventory.EnvironmentID.String() != payload.ScopeID {
		return fmt.Errorf("V1 environment adoption authority differs")
	}
	want, err := Evaluate(value.Inventory)
	if err != nil || !reflect.DeepEqual(value.Plan, want) {
		return fmt.Errorf("V1 environment adoption plan differs")
	}
	if !validAuthorityIDs(value.AuthorityIDs, value.Plan.OwnerUserID) {
		return fmt.Errorf("V1 environment adoption IAM identities are invalid")
	}
	return nil
}

func validAuthorityIDs(value AuthorityIDs, owner uuid.UUID) bool {
	ids := []uuid.UUID{
		value.SubjectID, value.OwnerRoleID,
		value.RolePermissionID, value.OwnerBindingID,
	}
	seen := map[uuid.UUID]struct{}{owner: {}}
	for _, id := range ids {
		if id == uuid.Nil {
			return false
		}
		if _, duplicate := seen[id]; duplicate {
			return false
		}
		seen[id] = struct{}{}
	}
	return true
}
