package postgres

import (
	"errors"
	"testing"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func TestBindWorkerFailsClosedWithoutWorkerPool(t *testing.T) {
	t.Parallel()
	store, err := BindWorker(nil, nil)
	if store != nil || !errors.Is(err, platformpostgres.ErrPoolClosed) {
		t.Fatalf("nil worker pool binding = %#v, %v", store, err)
	}
}
