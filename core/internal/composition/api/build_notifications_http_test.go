package api

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestNotificationsHTTPCompositionFailsClosedWithoutDependencies(t *testing.T) {
	if _, err := buildNotificationsHTTP(config.API{}, nil,
		platformasync.WriteCatalogs{}, nil, nil); err == nil {
		t.Fatal("Notifications HTTP composition accepted missing dependencies")
	}
}
