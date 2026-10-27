package postgres

import (
	"errors"
	"testing"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func TestBindRetentionFailsClosedWithoutWorkerPool(t *testing.T) {
	store, err := BindRetention(nil)
	if store != nil || !errors.Is(err, platformpostgres.ErrPoolClosed) {
		t.Fatalf("nil retention pool binding = %#v, %v", store, err)
	}
}
