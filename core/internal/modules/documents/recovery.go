package documents

import (
	"context"
	"fmt"

	documentspostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/recovery"
	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	recoveryapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/recovery"
	"github.com/google/uuid"
)

type recoveryFacade struct {
	service *recoveryapp.Service
}

func NewRecovery(
	database any,
	identity documentsapi.RecoveryIdentity,
) (documentsapi.Recovery, error) {
	environmentID, environmentErr := uuid.Parse(identity.EnvironmentID)
	tenantID, tenantErr := uuid.Parse(identity.TenantID)
	if environmentErr != nil || tenantErr != nil {
		return nil, fmt.Errorf("documents recovery identity is invalid")
	}
	internalIdentity := recoveryapp.Identity{
		EnvironmentID: environmentID, TenantID: tenantID,
		TenantSetSHA256: identity.TenantSetSHA256,
	}
	store, err := documentspostgres.Bind(database, internalIdentity)
	if err != nil {
		return nil, err
	}
	service, err := recoveryapp.New(store, internalIdentity)
	if err != nil {
		return nil, err
	}
	return &recoveryFacade{service: service}, nil
}

func (facade *recoveryFacade) BeginHold(
	ctx context.Context, recoverySetID string,
) error {
	setID, err := parseRecoveryID(recoverySetID)
	if err != nil {
		return err
	}
	return facade.service.BeginHold(ctx, setID)
}

func (facade *recoveryFacade) ReadInventoryAtSnapshot(
	ctx context.Context, recoverySetID, snapshotID string,
) ([]documentsapi.RecoveryInventoryItem, [32]byte, error) {
	setID, parseErr := parseRecoveryID(recoverySetID)
	if parseErr != nil {
		return nil, [32]byte{}, parseErr
	}
	items, digest, err := facade.service.ReadInventoryAtSnapshot(
		ctx, setID, snapshotID,
	)
	return exportRecoveryInventory(items), digest, err
}

func (facade *recoveryFacade) ListInventory(
	ctx context.Context, recoverySetID string,
) ([]documentsapi.RecoveryInventoryItem, [32]byte, error) {
	setID, parseErr := parseRecoveryID(recoverySetID)
	if parseErr != nil {
		return nil, [32]byte{}, parseErr
	}
	items, digest, err := facade.service.ListInventory(
		ctx, setID,
	)
	return exportRecoveryInventory(items), digest, err
}

func (facade *recoveryFacade) RecordCopy(
	ctx context.Context,
	recoverySetID string,
	item documentsapi.RecoveryInventoryItem,
	targetVersion string,
) error {
	setID, err := parseRecoveryID(recoverySetID)
	if err != nil {
		return err
	}
	internal, err := importRecoveryInventoryItem(item)
	if err != nil {
		return err
	}
	return facade.service.RecordCopy(
		ctx, setID, internal, targetVersion,
	)
}

func (facade *recoveryFacade) CompleteHold(
	ctx context.Context, recoverySetID string, digest [32]byte, count int64,
) error {
	setID, err := parseRecoveryID(recoverySetID)
	if err != nil {
		return err
	}
	return facade.service.CompleteHold(
		ctx, setID, digest, count,
	)
}

func (facade *recoveryFacade) ReleaseHold(
	ctx context.Context, recoverySetID string, digest [32]byte,
) error {
	setID, err := parseRecoveryID(recoverySetID)
	if err != nil {
		return err
	}
	return facade.service.ReleaseHold(
		ctx, setID, digest,
	)
}

func (facade *recoveryFacade) FailHold(
	ctx context.Context, recoverySetID, code string,
) error {
	setID, err := parseRecoveryID(recoverySetID)
	if err != nil {
		return err
	}
	return facade.service.FailHold(ctx, setID, code)
}

func (facade *recoveryFacade) ApplyLocator(
	ctx context.Context, remap documentsapi.RecoveryLocatorRemap,
) error {
	operationID, operationErr := parseRecoveryID(remap.RestoreOperationID)
	setID, setErr := parseRecoveryID(remap.RecoverySetID)
	if operationErr != nil || setErr != nil {
		return fmt.Errorf("documents recovery identity is invalid")
	}
	item, err := importRecoveryInventoryItem(remap.Item)
	if err != nil {
		return err
	}
	return facade.service.ApplyLocator(ctx, recoveryapp.LocatorRemap{
		RestoreOperationID:     operationID,
		RecoverySetID:          setID,
		Item:                   item,
		TargetObjectKey:        remap.TargetObjectKey,
		TargetObjectVersion:    remap.TargetObjectVersion,
		TargetObjectGeneration: remap.TargetObjectGeneration,
	})
}

func (facade *recoveryFacade) TerminalizeTransients(
	ctx context.Context, restoreOperationID string,
) (documentsapi.RecoveryTransientDisposition, error) {
	operationID, parseErr := parseRecoveryID(restoreOperationID)
	if parseErr != nil {
		return documentsapi.RecoveryTransientDisposition{}, parseErr
	}
	result, err := facade.service.TerminalizeTransients(
		ctx, operationID,
	)
	return documentsapi.RecoveryTransientDisposition{
		CancelledUploads:  result.CancelledUploads,
		ExpiredArchives:   result.ExpiredArchives,
		CancelledArchives: result.CancelledArchives,
	}, err
}

func parseRecoveryID(value string) (uuid.UUID, error) {
	identity, err := uuid.Parse(value)
	if err != nil || identity == uuid.Nil || identity.String() != value {
		return uuid.Nil, fmt.Errorf("documents recovery identity is invalid")
	}
	return identity, nil
}
