package worker

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildNotificationsDeliveryProcessors(settings config.Worker,
	release platformasync.ReleaseCatalogs, store asyncpostgres.WorkerStore,
	coordinator *platformtx.Coordinator,
) ([]jobexecution.Processor, error) {
	keyring, err := settings.NotificationsKEKKeyring()
	if err != nil {
		return nil, fmt.Errorf("load Notifications KEK keyring: %w", err)
	}
	keys, err := keyring.Keys()
	if err != nil {
		return nil, fmt.Errorf("read Notifications KEK keys: %w", err)
	}
	processors, err := notifications.NewProductionDeliveryJobProcessors(
		release, store, coordinator, keys)
	if err != nil {
		return nil, fmt.Errorf("build Notifications delivery processors: %w", err)
	}
	return processors, nil
}
