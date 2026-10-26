package catalog

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func TestNewAsyncPortFailsClosedWithoutReleaseCatalog(t *testing.T) {
	if port, err := NewAsyncPort(platformasync.WriteCatalogs{}, nil); err == nil || port != nil {
		t.Fatalf("port = %#v error = %v", port, err)
	}
}
