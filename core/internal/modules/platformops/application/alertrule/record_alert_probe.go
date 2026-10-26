package alertrule

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *AlertRuleOperations) RecordAlertTest(
	ctx context.Context, environmentID, id, actorID string,
) (platformopsapi.AlertEvent, error) {
	return operations.backend.AlertTestRecording(ctx, environmentID, id, actorID)
}
