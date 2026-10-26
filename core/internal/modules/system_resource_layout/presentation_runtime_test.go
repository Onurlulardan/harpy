package systemresourcelayout

import (
	"testing"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestNewPresentationRuntimeRequiresClock(t *testing.T) {
	if _, err := NewPresentationRuntime(nil); err != layoutapi.ErrInvalid {
		t.Fatalf("nil clock error = %v", err)
	}
	runtime, err := NewPresentationRuntime(func() time.Time {
		return time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC)
	})
	if err != nil || runtime == nil {
		t.Fatalf("presentation runtime = %#v: %v", runtime, err)
	}
}
