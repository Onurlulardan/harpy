package subscriptionapp

import "context"

// Store owns the final transaction boundary. Each method must acquire the
// operation receipt and subscription locks, invoke the planner once, enforce the
// planner's revision CAS, and atomically persist its receipt, audit, outbox and
// job changes. Planner error or commit failure leaves every effect absent.
//
// Actor is bound to the freshly authorized actor tenant. Provider resolves the
// tenant only from the verified durable provider mapping. Job revalidates the
// persisted tenant, billing+forms generations, lease owner and lease generation.
type Store interface {
	Actor(context.Context, ActorExecution, Planner) (Result, error)
	Provider(context.Context, ProviderExecution, Planner) (Result, error)
	Job(context.Context, JobExecution, Planner) (Result, error)
}

type Planner func(LockedState) (CommitPlan, error)
