package subscriptionapp

import (
	"github.com/google/uuid"
)

type Execution struct {
	OperationID       string
	SubscriptionID    uuid.UUID
	SourceOperationID uuid.UUID
	ExpectedRevision  int64
	RequestSHA256     [32]byte
}

type ActorExecution struct {
	Execution
	ActorID                 uuid.UUID
	Permission              string
	TargetOfferingVersionID uuid.UUID
	TargetPriceVersionID    uuid.UUID
}

type ProviderExecution struct {
	Execution
	ProviderReceiptID uuid.UUID
	NormalizedFactID  uuid.UUID
	NormalizedFactSHA [32]byte
}

type JobExecution struct {
	Execution
	JobID           uuid.UUID
	LeaseGeneration int64
	Boundary        BoundaryKind
}

func (value Execution) valid(operation string, expectedRequired bool) bool {
	return value.OperationID == operation && validID(value.SubscriptionID) &&
		validV7(value.SourceOperationID) && validDigest(value.RequestSHA256) &&
		(!expectedRequired || value.ExpectedRevision > 0)
}

func (value ActorExecution) valid(operation string) bool {
	permission := permissionBillingManage
	if operation == OperationPlatformTransition {
		permission = permissionPlatformManage
	}
	return value.Execution.valid(operation, true) && validID(value.ActorID) &&
		value.Permission == permission
}

func (value ProviderExecution) valid() bool {
	return value.Execution.valid(OperationUpdateFromProvider, false) &&
		validV7(value.ProviderReceiptID) && validV7(value.NormalizedFactID) &&
		validDigest(value.NormalizedFactSHA) &&
		value.SourceOperationID == value.ProviderReceiptID
}

func (value JobExecution) valid() bool {
	return value.Execution.valid(OperationExecuteDue, true) && validV7(value.JobID) &&
		value.LeaseGeneration > 0 && value.Boundary.valid()
}
