package cataloglifecycle

import (
	"context"
	"fmt"
	"time"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	catalogownerapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/catalogcomponentowner"
	"github.com/google/uuid"
)

func (store *Store) scanAuthority(
	row Row,
) (catalogownerapp.PurgeAuthorityBinding, error) {
	var result catalogownerapp.PurgeAuthorityBinding
	var operationID, componentID, resourceID string
	var versionID, headID string
	err := row.Scan(
		&operationID,
		&componentID,
		&result.ReferenceEpoch,
		&result.AuthoritySHA256,
		&resourceID,
		&versionID,
		&headID,
		&result.RecordedAt,
	)
	result.OperationID, err = parseUUID(operationID, err)
	result.ComponentID, err = parseUUID(componentID, err)
	result.ResourceID, err = parseUUID(resourceID, err)
	result.VersionID, err = parseUUID(versionID, err)
	result.HeadID, err = parseUUID(headID, err)
	if err != nil || !validAuthority(result) {
		return catalogownerapp.PurgeAuthorityBinding{}, unavailable()
	}
	return result, nil
}

func (store *Store) usable(
	ctx context.Context,
	component contract.Component,
) bool {
	return store != nil && store.queryRow != nil &&
		store.allowed != nil && store.allowed(ctx) &&
		component.Owner == owner &&
		component.Generation == store.generation &&
		component.ID != "" && component.Contract != "" &&
		digestPattern.MatchString(component.ContentSHA256)
}

func validAuthority(
	value catalogownerapp.PurgeAuthorityBinding,
) bool {
	return uuidV7(value.OperationID) &&
		uuidV7(value.ComponentID) &&
		value.ReferenceEpoch > 0 &&
		digestPattern.MatchString(value.AuthoritySHA256) &&
		uuidV7(value.ResourceID) &&
		uuidV7(value.VersionID) &&
		uuidV7(value.HeadID) &&
		validTime(value.RecordedAt)
}

func parseUUID(raw string, prior error) (uuid.UUID, error) {
	if prior != nil {
		return uuid.Nil, prior
	}
	value, err := uuid.Parse(raw)
	if err != nil || !uuidV7(value) {
		return uuid.Nil, unavailable()
	}
	return value, nil
}

func uuidV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 &&
		value.Variant() == uuid.RFC4122
}

func validTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0
}

func unavailable() error {
	return fmt.Errorf(
		"Analytics Catalog lifecycle store is unavailable",
	)
}
