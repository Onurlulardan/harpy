package platformops

import (
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type recoverySetView struct {
	RecoverySetID    string     `json:"recovery_set_id"`
	OperationID      string     `json:"operation_id"`
	EnvironmentID    string     `json:"environment_id"`
	Sequence         int64      `json:"sequence"`
	State            string     `json:"state"`
	ManifestSHA256   *string    `json:"manifest_sha256"`
	LegalHold        bool       `json:"legal_hold"`
	DatabaseSystemID *string    `json:"database_system_id"`
	ProtectedAt      *time.Time `json:"protected_at"`
}

func representRecoverySet(value platformopsapi.RecoverySet) recoverySetView {
	return recoverySetView{RecoverySetID: value.RecoverySetID,
		OperationID: value.OperationID, EnvironmentID: value.EnvironmentID,
		Sequence: value.Sequence, State: value.State,
		ManifestSHA256: optionalString(value.ManifestSHA256), LegalHold: value.LegalHold,
		DatabaseSystemID: optionalString(value.DatabaseSystemID), ProtectedAt: value.ProtectedAt}
}

type restoreOperationView struct {
	OperationID         string     `json:"operation_id"`
	RecoverySetID       string     `json:"recovery_set_id"`
	EnvironmentID       string     `json:"environment_id"`
	Disposition         string     `json:"disposition"`
	State               string     `json:"state"`
	PlanSHA256          *string    `json:"plan_sha256"`
	ReceiptSHA256       *string    `json:"receipt_sha256"`
	NewDatabaseSystemID *string    `json:"new_database_system_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	CompletedAt         *time.Time `json:"completed_at"`
}

func representRestore(value platformopsapi.RestoreOperation) restoreOperationView {
	return restoreOperationView{OperationID: value.OperationID,
		RecoverySetID: value.RecoverySetID, EnvironmentID: value.EnvironmentID,
		Disposition: string(value.Disposition), State: value.State,
		PlanSHA256:          optionalString(value.PlanSHA256),
		ReceiptSHA256:       optionalString(value.ReceiptSHA256),
		NewDatabaseSystemID: optionalString(value.NewDatabaseSystemID),
		CreatedAt:           value.CreatedAt, UpdatedAt: value.UpdatedAt, CompletedAt: value.CompletedAt}
}
