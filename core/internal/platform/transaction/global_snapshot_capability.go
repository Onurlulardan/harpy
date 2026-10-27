package transaction

import (
	"fmt"

	"github.com/google/uuid"
)

// GlobalSnapshotCapability proves an authenticated, freshly authorized actor
// may execute one read-only global Billing operation.
type GlobalSnapshotCapability struct {
	actorID              uuid.UUID
	operationID          string
	permission           string
	authorizationClock   int64
	authorizationVersion int64
	groupGeneration      int64
	billingGeneration    int64
	issued               bool
}

type GlobalSnapshotCapabilityIssuer struct{ valid bool }

func NewGlobalSnapshotCapabilityIssuer() GlobalSnapshotCapabilityIssuer {
	return GlobalSnapshotCapabilityIssuer{valid: true}
}

func (issuer GlobalSnapshotCapabilityIssuer) Issue(
	actorID uuid.UUID,
	operationID, permission string,
	authorizationClock, authorizationVersion int64,
	groupGeneration, billingGeneration int64,
) (GlobalSnapshotCapability, error) {
	capability := GlobalSnapshotCapability{
		actorID: actorID, operationID: operationID, permission: permission,
		authorizationClock: authorizationClock, authorizationVersion: authorizationVersion,
		groupGeneration: groupGeneration, billingGeneration: billingGeneration, issued: true,
	}
	if !issuer.valid || !capability.valid() {
		return GlobalSnapshotCapability{}, fmt.Errorf("global snapshot capability is invalid")
	}
	return capability, nil
}

func (capability GlobalSnapshotCapability) valid() bool {
	return capability.issued && capability.actorID != uuid.Nil &&
		operationIDPattern.MatchString(capability.operationID) &&
		validGlobalBillingPermission(capability.permission) &&
		capability.authorizationClock > 0 && capability.authorizationVersion > 0 &&
		capability.groupGeneration > 0 && capability.billingGeneration > 0
}

func (capability GlobalSnapshotCapability) Allows(
	actorID uuid.UUID,
	operationID, permission string,
	authorizationClock, authorizationVersion int64,
) bool {
	return capability.valid() && capability.actorID == actorID &&
		capability.operationID == operationID && capability.permission == permission &&
		capability.authorizationClock == authorizationClock &&
		capability.authorizationVersion == authorizationVersion
}
