package application

import "testing"

func TestSignalLegacyConnectionPersistsConsumerDigest(t *testing.T) {
	application, _, store, scope := realtimeFixture(t)
	digest := [32]byte{1, 2, 3}
	if err := application.SignalLegacyConnection(t.Context(), scope, digest); err != nil {
		t.Fatalf("SignalLegacyConnection() error = %v", err)
	}
	if store.hit.ConsumerManifestDigest != digest || !store.hit.TenantID.Valid() ||
		!store.hit.UserID.Valid() {
		t.Fatalf("legacy connection hit = %+v", store.hit)
	}
}
