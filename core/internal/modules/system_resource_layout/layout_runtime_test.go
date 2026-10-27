package systemresourcelayout

import (
	"testing"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func TestNewLayoutRuntimeRequiresFinalDependencies(t *testing.T) {
	clock := func() time.Time { return time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC) }
	ids := func() (uuid.UUID, error) { return uuid.NewV7() }
	if _, err := NewLayoutRuntime(nil, clock); err != layoutapi.ErrInvalid {
		t.Fatalf("nil IDs error = %v", err)
	}
	if _, err := NewLayoutRuntime(ids, nil); err != layoutapi.ErrInvalid {
		t.Fatalf("nil clock error = %v", err)
	}
	runtime, err := NewLayoutRuntime(ids, clock)
	if err != nil || runtime == nil {
		t.Fatalf("layout runtime = %#v: %v", runtime, err)
	}
}

func TestNewProductionLayoutRuntimeUsesUUIDv7(t *testing.T) {
	runtime, err := NewProductionLayoutRuntime(time.Now)
	if err != nil || runtime == nil {
		t.Fatalf("production runtime = %#v: %v", runtime, err)
	}
	id, err := runtime.ids.NewUUID()
	if err != nil || id.Version() != 7 {
		t.Fatalf("production ID = %s version=%d: %v", id, id.Version(), err)
	}
}
