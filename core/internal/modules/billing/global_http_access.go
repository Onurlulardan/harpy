package billing

import (
	"errors"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func snapshotCapability(
	access billingapi.Access, operation string,
) (platformtx.GlobalSnapshotCapability, error) {
	if !access.ValidFor(billingapi.PermissionPlatformBillingRead) ||
		access.Scope() != billingapi.AccessScopeGlobal {
		return platformtx.GlobalSnapshotCapability{}, billingapi.ErrForbidden
	}
	capability, valid := access.GlobalSnapshotCapabilityFor(operation)
	if !valid {
		return platformtx.GlobalSnapshotCapability{}, billingapi.ErrForbidden
	}
	return capability, nil
}

func catalogReadbackCapability(
	access billingapi.Access,
	readOperation string,
) (platformtx.GlobalSnapshotCapability, error) {
	if access.ValidFor(billingapi.PermissionPlatformBillingRead) {
		return snapshotCapability(access, readOperation)
	}
	if !access.ValidFor(billingapi.PermissionPlatformBillingManage) ||
		access.Scope() != billingapi.AccessScopeGlobal {
		return platformtx.GlobalSnapshotCapability{}, billingapi.ErrForbidden
	}
	for _, operation := range catalogMutationOperations() {
		_, readback, valid := access.GlobalCommandCapabilitiesFor(operation)
		if valid {
			return readback, nil
		}
	}
	return platformtx.GlobalSnapshotCapability{}, billingapi.ErrForbidden
}

func catalogMutationOperations() []string {
	return []string{
		offeringCreateOperation, offeringPublishOperation, offeringRetireOperation,
		discountCreateOperation, discountPublishOperation, discountRetireOperation,
		campaignCreateOperation, campaignPublishOperation, campaignRetireOperation,
	}
}

func commandCapabilities(
	access billingapi.Access, operation string,
) (platformtx.GlobalCommandCapability, platformtx.GlobalSnapshotCapability, error) {
	if !access.ValidFor(billingapi.PermissionPlatformBillingManage) ||
		access.Scope() != billingapi.AccessScopeGlobal {
		return platformtx.GlobalCommandCapability{},
			platformtx.GlobalSnapshotCapability{}, billingapi.ErrForbidden
	}
	command, readback, valid := access.GlobalCommandCapabilitiesFor(operation)
	if !valid {
		return platformtx.GlobalCommandCapability{},
			platformtx.GlobalSnapshotCapability{}, billingapi.ErrForbidden
	}
	return command, readback, nil
}

func normalizeScopeError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, billingapi.ErrForbidden) || errors.Is(err, billingapi.ErrInvalid) ||
		errors.Is(err, billingapi.ErrNotFound) || errors.Is(err, billingapi.ErrConflict) ||
		errors.Is(err, billingapi.ErrRevisionStale) ||
		errors.Is(err, billingapi.ErrIdempotencyReused) {
		return err
	}
	return billingapi.ErrAuthorityUnavailable
}
