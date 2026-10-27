package alertevent

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ operation string }

func (backend *backendFixture) AlertEventListing(
	context.Context, string, platformopsapi.AlertEventFilter,
) (platformopsapi.AlertEventPage, error) {
	backend.operation = "list"
	return platformopsapi.AlertEventPage{Total: 1}, nil
}

func (backend *backendFixture) AlertEventLookup(
	context.Context, string, string,
) (platformopsapi.AlertEvent, error) {
	backend.operation = "get"
	return platformopsapi.AlertEvent{ID: "event"}, nil
}
