package profilepostgres

import (
	"context"
	_ "embed"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

//go:embed sql/tenant/reserve_preconversion.sql
var profilePreconversionReserveSQL string

func reserveProfilePreconversion(
	ctx context.Context,
	bound *boundStore,
	contract profileWriteContract,
	resource identitydomain.ProfileResource,
	resourceID uuid.UUID,
	current profileWriteResource,
	planned profileapp.Document,
	target profileapp.Document,
) error {
	if bound == nil || !contract.valid(resource) || !contract.migrating() ||
		resourceID == uuid.Nil || current.actualSchemaID == uuid.Nil ||
		current.actualSchemaID == contract.target.ID() || current.revision < 1 ||
		planned.SchemaID() != contract.target.ID() ||
		target.SchemaID() != contract.target.ID() {
		return identitydomain.ErrStorage
	}
	sourceSHA := current.document.SHA256()
	plannedSHA := planned.SHA256()
	targetSHA := target.SHA256()
	var pageNumber int
	if bound.database.QueryRow(
		ctx, profilePreconversionReserveSQL,
		contract.planID, string(resource), resourceID,
		current.actualSchemaID, current.revision, sourceSHA[:],
		plannedSHA[:], targetSHA[:],
	).Scan(&pageNumber) != nil || pageNumber < 1 {
		return identitydomain.ErrStorage
	}
	return nil
}
