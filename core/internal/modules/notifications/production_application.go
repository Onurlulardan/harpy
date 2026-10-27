package notifications

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func buildProductionNotificationApplication(
	catalogs platformasync.WriteCatalogs,
) (*postgres.Binder, *application.Application, error) {
	if httpNil(catalogs) {
		return nil, nil, application.ErrConfiguration
	}
	binder, err := newPlatformNotificationBinder(catalogs)
	if err != nil {
		return nil, nil, err
	}
	base, err := application.NewApplication(binder, recipientResolver{binder: binder},
		newScopeAuthorizer(), productionClock{}, productionIDs{})
	if err != nil {
		return nil, nil, fmt.Errorf("build notification application: %w", err)
	}
	return binder, base, nil
}
