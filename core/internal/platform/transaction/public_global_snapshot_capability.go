package transaction

import "fmt"

type PublicGlobalOperation string

const (
	PublicBillingCatalogGet  PublicGlobalOperation = "billing.catalog.get"
	PublicBillingCatalogList PublicGlobalOperation = "billing.catalog.list"
)

type PublicGlobalSnapshotCapability struct {
	operation         PublicGlobalOperation
	groupGeneration   int64
	billingGeneration int64
	issued            bool
}

type PublicGlobalSnapshotCapabilityIssuer struct {
	valid bool
}

func NewPublicGlobalSnapshotCapabilityIssuer() PublicGlobalSnapshotCapabilityIssuer {
	return PublicGlobalSnapshotCapabilityIssuer{valid: true}
}

func (issuer PublicGlobalSnapshotCapabilityIssuer) Issue(
	operation PublicGlobalOperation,
	groupGeneration int64,
	billingGeneration int64,
) (PublicGlobalSnapshotCapability, error) {
	capability := PublicGlobalSnapshotCapability{
		operation: operation, groupGeneration: groupGeneration,
		billingGeneration: billingGeneration, issued: true,
	}
	if !issuer.valid || !capability.valid() {
		return PublicGlobalSnapshotCapability{}, fmt.Errorf(
			"public global snapshot capability is invalid",
		)
	}
	return capability, nil
}

func (capability PublicGlobalSnapshotCapability) valid() bool {
	return capability.issued && capability.operation.valid() &&
		capability.groupGeneration > 0 && capability.billingGeneration > 0
}

func (operation PublicGlobalOperation) valid() bool {
	return operation == PublicBillingCatalogGet || operation == PublicBillingCatalogList
}
