package cataloglifecycle

import (
	"context"
	"encoding/json"
	"regexp"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	catalogownerapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/catalogcomponentowner"
	"github.com/google/uuid"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Row interface{ Scan(...any) error }

type Store struct {
	generation int64
	queryRow   func(context.Context, string, ...any) Row
	allowed    func(context.Context) bool
}

func (store *Store) Binding(
	ctx context.Context,
	component contract.Component,
	packageID uuid.UUID,
) (catalogownerapp.ResourceBinding, error) {
	if !store.usable(ctx, component) || !uuidV7(packageID) {
		return catalogownerapp.ResourceBinding{}, unavailable()
	}
	var result catalogownerapp.ResourceBinding
	var resourceID, versionID, headID string
	err := store.queryRow(
		ctx,
		bindingSQL,
		component.ID,
		packageID,
		store.generation,
		component.Contract,
		component.ContentSHA256,
	).Scan(
		&resourceID,
		&versionID,
		&headID,
		&result.Lifecycle,
		&result.ResourceReferenceEpoch,
	)
	result.ResourceID, err = parseUUID(resourceID, err)
	result.VersionID, err = parseUUID(versionID, err)
	result.HeadID, err = parseUUID(headID, err)
	if err != nil || result.Lifecycle == "" ||
		result.ResourceReferenceEpoch < 1 {
		return catalogownerapp.ResourceBinding{}, unavailable()
	}
	return result, nil
}

func (store *Store) Preflight(
	ctx context.Context,
	component contract.Component,
	packageID uuid.UUID,
	referenceEpoch uint64,
) (catalogownerapp.PurgeSnapshot, error) {
	if !store.usable(ctx, component) || !uuidV7(packageID) ||
		referenceEpoch == 0 {
		return catalogownerapp.PurgeSnapshot{}, unavailable()
	}
	var document []byte
	var resourceID, versionID, headID string
	err := store.queryRow(
		ctx,
		preflightSQL,
		component.ID,
		packageID,
		store.generation,
		component.Contract,
		component.ContentSHA256,
		referenceEpoch,
	).Scan(&document, &resourceID, &versionID, &headID)
	result := catalogownerapp.PurgeSnapshot{}
	result.ResourceID, err = parseUUID(resourceID, err)
	result.VersionID, err = parseUUID(versionID, err)
	result.HeadID, err = parseUUID(headID, err)
	if err != nil || json.Unmarshal(document, &result.Blockers) != nil ||
		len(result.Blockers) > 100 {
		return catalogownerapp.PurgeSnapshot{}, unavailable()
	}
	result.Lifecycle = "RETIRED"
	return result, nil
}

func (store *Store) PersistPurgeAuthority(
	ctx context.Context,
	component contract.Component,
	packageID uuid.UUID,
	value catalogownerapp.PurgeAuthorityBinding,
) (catalogownerapp.PurgeAuthorityBinding, error) {
	if !store.usable(ctx, component) || !uuidV7(packageID) ||
		!validAuthority(value) {
		return catalogownerapp.PurgeAuthorityBinding{}, unavailable()
	}
	return store.scanAuthority(
		store.queryRow(
			ctx,
			recordPurgeAuthoritySQL,
			value.OperationID,
			component.ID,
			packageID,
			store.generation,
			component.Contract,
			component.ContentSHA256,
			value.ReferenceEpoch,
			value.AuthoritySHA256,
			value.ResourceID,
			value.VersionID,
			value.HeadID,
			value.RecordedAt,
		),
	)
}

func (store *Store) LoadPurgeAuthority(
	ctx context.Context,
	component contract.Component,
	packageID uuid.UUID,
	operationID uuid.UUID,
) (catalogownerapp.PurgeAuthorityBinding, error) {
	if !store.usable(ctx, component) || !uuidV7(packageID) ||
		!uuidV7(operationID) {
		return catalogownerapp.PurgeAuthorityBinding{}, unavailable()
	}
	return store.scanAuthority(
		store.queryRow(
			ctx,
			readPurgeAuthoritySQL,
			operationID,
			component.ID,
			packageID,
			store.generation,
			component.Contract,
			component.ContentSHA256,
		),
	)
}

var _ catalogownerapp.LifecycleStore = (*Store)(nil)
