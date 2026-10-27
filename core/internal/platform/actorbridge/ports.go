package actorbridge

import (
	"context"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/google/uuid"
)

type IssueAuthorityPort interface {
	ResolveIssueAuthority(context.Context, uuid.UUID, actorassertionv1.Domain) (Authority, error)
}

type ConsumePort interface {
	Consume(context.Context, actorassertionv1.Claims) (Receipt, error)
}

type MaintenancePort interface {
	ReconcileHits(context.Context, int) (Reconciliation, error)
	Purge(context.Context, int) (int64, error)
}
