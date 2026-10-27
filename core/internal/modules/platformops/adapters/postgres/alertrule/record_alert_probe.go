package alertrule

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *AlertRuleOperations) RecordAlertTest(
	ctx context.Context, environmentID, id, actorID, message string,
) (platformopsapi.AlertEvent, error) {
	return operations.backend.AlertTestStorage(ctx, environmentID, id, actorID, message)
}
