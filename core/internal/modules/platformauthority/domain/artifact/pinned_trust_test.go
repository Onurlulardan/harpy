package artifact

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"
)

func TestPinnedTrustVerifiesAckWithoutReleaseAuthority(t *testing.T) {
	fixture := newArtifactFixture(t)
	bundle, err := SignTrustBundle(TrustBundle{Schema: TrustBundleSchema,
		RegistryKind: RegistryDevelopment, TrustGeneration: 1,
		Active: TrustKey{KID: "registry.v1", Algorithm: AlgorithmEd25519,
			PublicKey: base64.RawURLEncoding.EncodeToString(fixture.registryPublic),
			NotBefore: fixtureNow.Add(-time.Hour).Format(time.RFC3339),
			SignUntil: fixtureNow.Add(time.Hour).Format(time.RFC3339)},
		Retired: []RetiredTrustKey{}}, fixture.releasePrivate)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(bundle)
	pinned, err := VerifyPinnedTrustBundle(bundle, RegistryDevelopment,
		hex.EncodeToString(digest[:]), fixtureNow)
	if err != nil {
		t.Fatal(err)
	}
	ack := fixture.ack(t, fixture.registryPrivate)
	if _, err = VerifyAckPinned(ack.Bytes, pinned, fixtureNow); err != nil {
		t.Fatal(err)
	}
	if _, err = VerifyPinnedTrustBundle(bundle, RegistryDevelopment,
		hex.EncodeToString(make([]byte, 32)), fixtureNow); err == nil {
		t.Fatal("wrong trust digest was accepted")
	}
}
