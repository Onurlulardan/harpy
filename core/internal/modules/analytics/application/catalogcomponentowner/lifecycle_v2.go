package catalogcomponentowner

import (
	"context"
	"crypto/sha256"
	"fmt"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (owner *Owner) lifecycleStore(
	ctx context.Context,
	scope platformtx.TenantScope,
	tenantRaw string,
	packageRaw string,
	component componentowner.Component,
	valid bool,
) (LifecycleStore, uuid.UUID, error) {
	tenantID, tenantErr := uuid.Parse(tenantRaw)
	packageID, packageErr := uuid.Parse(packageRaw)
	if owner == nil || owner.lifecycle == nil || !valid ||
		component.Owner != "analytics" ||
		tenantErr != nil || packageErr != nil ||
		tenantID.Version() != 7 || packageID.Version() != 7 {
		return nil, uuid.Nil, fmt.Errorf(
			"Analytics Catalog lifecycle command is invalid",
		)
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "analytics")
	if err != nil || !authority.OwnsScope(scope) ||
		authority.ScopeID() != tenantID ||
		authority.Generation() != component.Generation {
		return nil, uuid.Nil, fmt.Errorf(
			"Analytics Catalog lifecycle scope is unavailable",
		)
	}
	store, err := owner.lifecycle.Bind(
		ctx,
		scope,
		component.Generation,
	)
	if err != nil || store == nil {
		return nil, uuid.Nil, fmt.Errorf(
			"Analytics Catalog lifecycle store is unavailable",
		)
	}
	return store, packageID, nil
}

func (owner *Owner) purgeStore(
	ctx context.Context,
	scope platformtx.TenantScope,
	command componentowner.PurgeCommand,
) (
	LifecycleStore,
	uuid.UUID,
	uuid.UUID,
	uuid.UUID,
	error,
) {
	store, packageID, err := owner.lifecycleStore(
		ctx,
		scope,
		command.Authority.TenantID,
		command.Authority.PackageID,
		command.Component,
		command.Valid(),
	)
	operationID, operationErr := uuid.Parse(
		command.Authority.OperationID,
	)
	componentID, componentErr := uuid.Parse(command.Component.ID)
	if err != nil || operationErr != nil || componentErr != nil ||
		operationID.Version() != 7 || componentID.Version() != 7 {
		return nil, uuid.Nil, uuid.Nil, uuid.Nil,
			fmt.Errorf("Analytics Catalog purge command is invalid")
	}
	return store, packageID, operationID, componentID, nil
}

func lifecycleResult(
	command componentowner.LifecycleCommand,
	receipt componentowner.Receipt,
	binding ResourceBinding,
) componentowner.OwnerResult {
	state := catalogcomponentv2.StateRetired
	if command.Action == "restore" {
		state = catalogcomponentv2.StateRestored
	}
	result := componentowner.OwnerResult{
		ReceiptID:      receipt.ID,
		TenantID:       command.TenantID,
		PackageID:      command.PackageID,
		Component:      command.Component,
		OperationID:    command.OperationID,
		ResourceID:     binding.ResourceID.String(),
		VersionID:      binding.VersionID.String(),
		HeadID:         binding.HeadID.String(),
		ReferenceEpoch: command.ReferenceEpoch,
		State:          state,
		Ready:          true,
		CompletedAt:    receipt.CompletedAt,
	}
	result.ResultSHA256 = catalogcomponentv2.ResultSHA256(result)
	return result
}

func purgeResult(
	command componentowner.PurgeCommand,
	receiptID uuid.UUID,
	binding ResourceBinding,
) componentowner.OwnerResult {
	return componentowner.OwnerResult{
		ReceiptID:       receiptID.String(),
		TenantID:        command.Authority.TenantID,
		PackageID:       command.Authority.PackageID,
		Component:       command.Component,
		OperationID:     command.Authority.OperationID,
		ResourceID:      binding.ResourceID.String(),
		VersionID:       binding.VersionID.String(),
		HeadID:          binding.HeadID.String(),
		ReferenceEpoch:  command.Authority.ReferenceEpoch,
		AuthoritySHA256: command.Authority.AuthoritySHA256,
	}
}

func legacyLifecycleCommand(
	command componentowner.LifecycleCommand,
) componentowner.Command {
	return componentowner.Command{
		OperationID:    command.OperationID,
		Component:      command.Component,
		Action:         command.Action,
		IdempotencyKey: command.IdempotencyKey,
	}
}

func authorityMatches(
	command componentowner.PurgeCommand,
	binding ResourceBinding,
	value PurgeAuthorityBinding,
) bool {
	return value.OperationID.String() ==
		command.Authority.OperationID &&
		value.ComponentID.String() == command.Component.ID &&
		value.ReferenceEpoch ==
			int64(command.Authority.ReferenceEpoch) &&
		value.AuthoritySHA256 ==
			command.Authority.AuthoritySHA256 &&
		value.ResourceID == binding.ResourceID &&
		value.VersionID == binding.VersionID &&
		value.HeadID == binding.HeadID
}

func deriveReceiptID(
	operationID uuid.UUID,
	componentID uuid.UUID,
	phase string,
) uuid.UUID {
	digest := sha256.Sum256(
		append(
			append(
				[]byte(
					"analytics.catalog-owner."+
						phase+
						".v2\x00",
				),
				operationID[:]...,
			),
			componentID[:]...,
		),
	)
	var result uuid.UUID
	copy(result[:6], operationID[:6])
	copy(result[6:], digest[6:16])
	result[6] = result[6]&0x0f | 0x70
	result[8] = result[8]&0x3f | 0x80
	return result
}

func unavailableOwnerResult() error {
	return fmt.Errorf("Analytics Catalog owner result is unavailable")
}

var _ componentowner.LifecycleOwner = (*Owner)(nil)
