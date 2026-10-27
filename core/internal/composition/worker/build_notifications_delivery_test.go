package worker

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestNotificationsDeliveryCompositionFailsClosedWithoutKeyring(t *testing.T) {
	release, err := platformasync.LoadReleaseCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = buildNotificationsDeliveryProcessors(config.Worker{}, release, nil, nil); err == nil {
		t.Fatal("Notifications delivery composition accepted an empty keyring")
	}
}
