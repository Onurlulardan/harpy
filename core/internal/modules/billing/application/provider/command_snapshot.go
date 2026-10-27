package providerapp

import (
	"time"

	"github.com/google/uuid"
)

type CommandSnapshot struct {
	CommandID, FinalizerJobID uuid.UUID
	TenantID, ActorID         uuid.UUID
	ProviderAccountID         uuid.UUID
	ProviderObjectMappingID   *uuid.UUID
	SubscriptionID            uuid.UUID
	Provider                  string
	Kind                      CommandKind
	EffectKey                 string
	RequestSHA256             [32]byte
	State                     CommandState
	RequestedAt               time.Time
	Revision                  int64
	Replay                    bool
}

func (snapshot CommandSnapshot) valid() bool {
	provider, err := CanonicalProviderName(snapshot.Provider)
	return err == nil && provider == snapshot.Provider &&
		snapshot.CommandID != uuid.Nil &&
		snapshot.CommandID.Version() == uuid.Version(7) &&
		snapshot.FinalizerJobID != uuid.Nil &&
		snapshot.FinalizerJobID.Version() == uuid.Version(7) &&
		snapshot.CommandID != snapshot.FinalizerJobID &&
		snapshot.TenantID != uuid.Nil && snapshot.ActorID != uuid.Nil &&
		snapshot.ProviderAccountID != uuid.Nil && snapshot.SubscriptionID != uuid.Nil &&
		optionalCommandID(snapshot.ProviderObjectMappingID) && snapshot.Kind.Valid() &&
		validCommandText(snapshot.EffectKey, 256) &&
		snapshot.RequestSHA256 != [32]byte{} && snapshot.State.Valid() &&
		!snapshot.RequestedAt.IsZero() && snapshot.RequestedAt.Location() == time.UTC &&
		snapshot.RequestedAt.Nanosecond()%int(time.Microsecond) == 0 &&
		snapshot.Revision > 0
}
