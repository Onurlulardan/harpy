package billing

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestNewPlatformReportsRuntimeRejectsMissingCoordinator(t *testing.T) {
	_, err := NewPlatformReportsRuntime(
		nil, config.CursorKeyring{}, platformasync.WriteCatalogs{},
	)
	if err == nil {
		t.Fatal("missing platform report coordinator was accepted")
	}
}
