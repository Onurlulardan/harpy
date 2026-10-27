package worker

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func buildNotificationsRetentionProcessors(
	release platformasync.ReleaseCatalogs,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	processors, err := notifications.NewRetentionJobProcessors(release, store)
	if err != nil {
		return nil, fmt.Errorf("build Notifications retention processors: %w", err)
	}
	return processors, nil
}
