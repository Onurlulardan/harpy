package postgres

import (
	"context"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	"github.com/jackc/pgx/v5"
)

type ActorBridgeQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type ActorBridgeDatabase struct {
	queryer ActorBridgeQueryer
}

func NewActorBridgeDatabase(queryer ActorBridgeQueryer) (*ActorBridgeDatabase, error) {
	if queryer == nil || actorBridgeQueryerIsNil(queryer) {
		return nil, actorbridge.NewFault(actorbridge.FaultInvalid)
	}
	return &ActorBridgeDatabase{queryer: queryer}, nil
}

func actorBridgeQueryerIsNil(queryer ActorBridgeQueryer) bool {
	value := reflect.ValueOf(queryer)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

var _ actorbridge.IssueAuthorityPort = (*ActorBridgeDatabase)(nil)
var _ actorbridge.MaintenancePort = (*ActorBridgeDatabase)(nil)
