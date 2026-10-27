package providerapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CommandStore interface {
	CreateCommand(context.Context, CommandMutation) (CommandSnapshot, error)
}

type CommandIDSource func() (uuid.UUID, error)
type CommandClock func() time.Time

type CommandMutation struct {
	CommandID, FinalizerJobID uuid.UUID
	TenantID, ActorID         uuid.UUID
	ProviderObjectMappingID   *uuid.UUID
	SubscriptionID            uuid.UUID
	Kind                      CommandKind
	EffectKey                 string
	RequestSHA256             [32]byte
	RequestedAt               time.Time
}
