package authorization

import (
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

type Requirement struct {
	id           [sha256.Size]byte
	kind         RequirementKind
	catalog      api.CatalogReference
	resourcePath resolved.RequestPath
	requestPath  resolved.RequestPath
	fieldID      uuid.UUID
	relationID   uuid.UUID
	validated    bool
}

func newRequirement(
	kind RequirementKind,
	catalog api.CatalogReference,
	resourcePath resolved.RequestPath,
	requestPath resolved.RequestPath,
	fieldID uuid.UUID,
	relationID uuid.UUID,
) (Requirement, error) {
	requirement := Requirement{
		kind: kind, catalog: catalog, resourcePath: resourcePath,
		requestPath: requestPath, fieldID: fieldID, relationID: relationID,
		validated: true,
	}
	identity, err := requirement.identitySHA256()
	if err != nil {
		return Requirement{}, err
	}
	requirement.id = identity
	if err := requirement.Validate(); err != nil {
		return Requirement{}, err
	}
	return requirement, nil
}

func (requirement Requirement) Validate() error {
	if !requirement.validated || requirement.kind.Validate() != nil ||
		requirement.catalog.Validate() != nil || requirement.resourcePath.Validate() != nil ||
		requirement.requestPath.Validate() != nil || !validRequirementTarget(requirement) ||
		!requestPathBelongsToResource(requirement.resourcePath, requirement.requestPath) {
		return fmt.Errorf("REST Query authorization requirement is invalid")
	}
	expected, err := requirement.identitySHA256()
	if err != nil || expected == [sha256.Size]byte{} || expected != requirement.id {
		return fmt.Errorf("REST Query authorization requirement is invalid")
	}
	return nil
}

func (requirement Requirement) ID() ([sha256.Size]byte, error) {
	if err := requirement.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return requirement.id, nil
}

func (requirement Requirement) Kind() (RequirementKind, error) {
	if err := requirement.Validate(); err != nil {
		return 0, err
	}
	return requirement.kind, nil
}

func (requirement Requirement) Catalog() (api.CatalogReference, error) {
	if err := requirement.Validate(); err != nil {
		return api.CatalogReference{}, err
	}
	return requirement.catalog, nil
}

func (requirement Requirement) ResourcePath() (resolved.RequestPath, error) {
	if err := requirement.Validate(); err != nil {
		return resolved.RequestPath{}, err
	}
	return requirement.resourcePath, nil
}

func (requirement Requirement) RequestPath() (resolved.RequestPath, error) {
	if err := requirement.Validate(); err != nil {
		return resolved.RequestPath{}, err
	}
	return requirement.requestPath, nil
}

func (requirement Requirement) FieldID() (uuid.UUID, bool, error) {
	if err := requirement.Validate(); err != nil {
		return uuid.Nil, false, err
	}
	return requirement.fieldID, requirement.fieldID != uuid.Nil, nil
}

func (requirement Requirement) RelationID() (uuid.UUID, bool, error) {
	if err := requirement.Validate(); err != nil {
		return uuid.Nil, false, err
	}
	return requirement.relationID, requirement.relationID != uuid.Nil, nil
}

func (Requirement) String() string   { return authorizationRedaction }
func (Requirement) GoString() string { return authorizationRedaction }
func (Requirement) LogValue() slog.Value {
	return authorizationLogValue()
}
func (requirement Requirement) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (Requirement) MarshalJSON() ([]byte, error) { return rejectAuthorizationJSON() }
