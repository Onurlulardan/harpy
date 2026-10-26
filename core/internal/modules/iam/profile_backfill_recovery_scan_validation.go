package iam

import (
	"bytes"

	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	"github.com/google/uuid"
)

func validRecoveryInventoryRows(
	contract profilebackfillpostgres.RecoveryContract,
	after uuid.UUID,
	rows []profilebackfillpostgres.RecoveryInventoryRow,
) bool {
	previous := after
	for _, row := range rows {
		if row.Validate() != nil || previous != uuid.Nil &&
			bytes.Compare(previous[:], row.Resource.ID[:]) >= 0 ||
			!validRecoveryInventoryLineage(contract, row) {
			return false
		}
		previous = row.Resource.ID
	}
	return true
}

func validRecoveryInventoryLineage(
	contract profilebackfillpostgres.RecoveryContract,
	row profilebackfillpostgres.RecoveryInventoryRow,
) bool {
	if row.LineageKind == "target" {
		return row.Resource.SchemaVersionID == contract.Target.ID() &&
			row.SchemaVersion == contract.Target.Version() && row.SchemaState == "draft"
	}
	if row.LineageKind != "source" || row.SchemaVersion > contract.Source.Version() {
		return false
	}
	if row.SchemaState == "active" {
		return row.Resource.SchemaVersionID == contract.Source.ID() &&
			row.SchemaVersion == contract.Source.Version()
	}
	return row.SchemaState == "retired" &&
		row.Resource.SchemaVersionID != contract.Target.ID()
}
