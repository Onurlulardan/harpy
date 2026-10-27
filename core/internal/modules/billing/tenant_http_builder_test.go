package billing

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func TestNewTenantHTTPRuntimeFailsClosedWithoutPlatformBindings(t *testing.T) {
	if _, err := NewTenantHTTPRuntime(nil, nil, platformasync.WriteCatalogs{}); err == nil {
		t.Fatal("tenant Billing production runtime accepted missing platform bindings")
	}
}
