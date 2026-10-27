package compatibility

import (
	"context"
	"time"
)

type AdoptionStore interface {
	BackfillItem(context.Context, AdoptionAuthority, AdoptionItem) (AdoptionItem, bool, error)
	Accept(context.Context, AdoptionAuthority, Acceptance) (Acceptance, bool, error)
}

type PublicLegacyStore interface {
	Resolve(context.Context, LegacyKey) (LegacyMapping, error)
	RecordHit(context.Context, LegacyMapping, time.Time) (LegacyMapping, error)
}

type LegacyStore interface {
	PublicLegacyStore
	FinalizeConsumers(context.Context, LegacyMapping, [32]byte, time.Time) (LegacyMapping, error)
	Retire(context.Context, LegacyMapping, RetirementEvidence, time.Time) (LegacyMapping, error)
}

type RetirementAuthority interface {
	Evidence(context.Context) (RetirementEvidence, error)
}

type Clock interface{ Now() time.Time }
