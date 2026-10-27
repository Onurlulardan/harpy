package ports

import "context"

type DeliveryClaim struct {
	OperationID     string
	LeaseOwnerID    string
	LeaseGeneration int64
	Attempt         int
	MaxAttempts     int
}

type DeliveryStore interface {
	ClaimDelivery(context.Context, string) (DeliveryClaim, bool, error)
	RetryDelivery(context.Context, DeliveryClaim, string, bool) error
	UpdateDeliveryImport(context.Context, DeliveryClaim) error
	CompleteDelivery(context.Context, DeliveryClaim) error
}

type OperationFinalizer interface {
	FinalizeAuthorityOperation(context.Context, OperationRecord) error
}
