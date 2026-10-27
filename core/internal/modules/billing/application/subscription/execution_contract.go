package subscriptionapp

import "github.com/google/uuid"

// ActorExecutionAccepted is the storage-boundary contract for actor commands.
func ActorExecutionAccepted(value ActorExecution) bool {
	if !value.valid(value.OperationID) {
		return false
	}
	switch value.OperationID {
	case OperationCancelAtPeriodEnd, OperationPlatformTransition,
		OperationResumeCancellation:
		return value.TargetOfferingVersionID == uuid.Nil &&
			value.TargetPriceVersionID == uuid.Nil
	case OperationChangeAtPeriodEnd:
		return validID(value.TargetOfferingVersionID) &&
			validID(value.TargetPriceVersionID)
	case OperationClaimTrial:
		return validID(value.TargetOfferingVersionID) &&
			value.TargetPriceVersionID == uuid.Nil
	default:
		return false
	}
}

// ProviderExecutionAccepted is the storage-boundary contract for provider facts.
func ProviderExecutionAccepted(value ProviderExecution) bool { return value.valid() }

// JobExecutionAccepted is the storage-boundary contract for leased jobs.
func JobExecutionAccepted(value JobExecution) bool { return value.valid() }
