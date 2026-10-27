package postgres

import "testing"

func TestBindPublisherStoreRequiresWorkerPool(t *testing.T) {
	t.Parallel()
	if store, err := BindPublisherStore(nil); err == nil || store != nil {
		t.Fatalf("nil publisher pool = %#v, %v", store, err)
	}
}
