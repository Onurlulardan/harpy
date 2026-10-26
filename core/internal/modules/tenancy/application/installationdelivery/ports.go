package installationdelivery

import "context"

type Store interface {
	Stage(context.Context, StageCommand) (Delivery, error)
	Read(context.Context, ReadQuery) (Delivery, error)
	Claim(context.Context, ClaimCommand) (Delivery, error)
	Acknowledge(context.Context, AcknowledgeCommand) (Delivery, error)
	RecordUnknownOutcome(context.Context, UnknownCommand) (Transition, error)
	DeleteSecret(context.Context, StateCommand) (Transition, error)
	Complete(context.Context, StateCommand) (Transition, error)
}
