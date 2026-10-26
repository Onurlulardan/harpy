package workhttp

import (
	"context"
	"encoding/json"

	restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type ListInput struct {
	Access        Access
	Operation     string
	Resource      string
	IDs           map[string]string
	Query         restapi.QueryRequest
	LegacyQuery   map[string]string
	Compatibility bool
}

type GetInput struct {
	Access        Access
	Operation     string
	Resource      string
	IDs           map[string]string
	Compatibility bool
}

type WriteInput struct {
	Access         Access
	Resource       string
	Operation      string
	IDs            map[string]string
	Body           json.RawMessage
	IdempotencyKey string
	Compatibility  bool
}

type CompatibilityHit struct {
	Access        Access
	RouteContract string
	ConsumerID    string
	BuildID       string
}

type Application interface {
	List(context.Context, ListInput) (json.RawMessage, error)
	Get(context.Context, GetInput) (json.RawMessage, error)
	Write(context.Context, WriteInput) (json.RawMessage, error)
	RecordCompatibilityHit(context.Context, CompatibilityHit) error
}
