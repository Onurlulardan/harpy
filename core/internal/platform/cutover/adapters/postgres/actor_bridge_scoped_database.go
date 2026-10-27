package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
)

type ActorBridgeRow interface {
	Scan(...any) error
}

type ActorBridgeQueryRow func(
	context.Context,
	string,
	...any,
) ActorBridgeRow

type ActorBridgeIssueDatabase struct {
	queryRow ActorBridgeQueryRow
}

func NewActorBridgeIssueDatabase(
	queryRow ActorBridgeQueryRow,
) (*ActorBridgeIssueDatabase, error) {
	if queryRow == nil {
		return nil, actorbridge.NewFault(actorbridge.FaultInvalid)
	}
	return &ActorBridgeIssueDatabase{queryRow: queryRow}, nil
}

var _ actorbridge.IssueAuthorityPort = (*ActorBridgeIssueDatabase)(nil)
