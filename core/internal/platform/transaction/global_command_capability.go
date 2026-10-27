package transaction

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const globalBillingPermissionPrefix = "platform.billing."

// GlobalCommandCapability is an opaque proof of an authenticated actor and a
// fresh IAM authorization decision for one global Billing command.
type GlobalCommandCapability struct {
	actorID              uuid.UUID
	operationID          string
	permission           string
	authorizationClock   int64
	authorizationVersion int64
	groupGeneration      int64
	billingGeneration    int64
	issued               bool
}

type GlobalCommandCapabilityIssuer struct {
	valid bool
}

func NewGlobalCommandCapabilityIssuer() GlobalCommandCapabilityIssuer {
	return GlobalCommandCapabilityIssuer{valid: true}
}

func (issuer GlobalCommandCapabilityIssuer) Issue(
	actorID uuid.UUID,
	operationID string,
	permission string,
	authorizationClock int64,
	authorizationVersion int64,
	groupGeneration int64,
	billingGeneration int64,
) (GlobalCommandCapability, error) {
	capability := GlobalCommandCapability{
		actorID: actorID, operationID: operationID, permission: permission,
		authorizationClock:   authorizationClock,
		authorizationVersion: authorizationVersion,
		groupGeneration:      groupGeneration, billingGeneration: billingGeneration,
		issued: true,
	}
	if !issuer.valid || !capability.valid() {
		return GlobalCommandCapability{}, fmt.Errorf("global command capability is invalid")
	}
	return capability, nil
}

func (capability GlobalCommandCapability) valid() bool {
	return capability.issued && capability.actorID != uuid.Nil &&
		operationIDPattern.MatchString(capability.operationID) &&
		validGlobalBillingPermission(capability.permission) &&
		capability.authorizationClock > 0 && capability.authorizationVersion > 0 &&
		capability.groupGeneration > 0 && capability.billingGeneration > 0
}

func validGlobalBillingPermission(permission string) bool {
	if !strings.HasPrefix(permission, globalBillingPermissionPrefix) {
		return false
	}
	action := strings.TrimPrefix(permission, globalBillingPermissionPrefix)
	return operationIDPattern.MatchString(action)
}
