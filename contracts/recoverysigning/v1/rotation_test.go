package recoverysigningv1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func TestRotateAdvancesGenerationAndRetainsEveryOldKeyVerifyOnly(t *testing.T) {
	created := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	first, err := Build("018f0000-0000-7000-8000-000000000001", created,
		fixtureMaterial())
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(first.CanonicalBytes(), []byte(`"retired"`)) {
		t.Fatal("generation-1 wire compatibility drifted")
	}
	second, err := Rotate(first, created.Add(time.Hour), rotatedMaterial(4))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(second.CanonicalBytes())
	if err != nil || parsed.Generation() != 2 {
		t.Fatalf("rotated bundle parse = generation %d, err=%v", parsed.Generation(), err)
	}
	for _, purpose := range orderedPurposes {
		oldID, _ := first.KeyID(purpose)
		oldPublic, _ := first.PublicKey(purpose)
		digest := sha256.Sum256(oldPublic[:])
		if !parsed.Retains(purpose, oldID, hex.EncodeToString(digest[:])) {
			t.Fatalf("old %q key is not retained", purpose)
		}
	}
	third, err := Rotate(second, created.Add(2*time.Hour), rotatedMaterial(7))
	if err != nil {
		t.Fatal(err)
	}
	if parsedThird, err := Parse(third.CanonicalBytes()); err != nil ||
		parsedThird.Generation() != 3 || len(parsedThird.retired) != 6 {
		t.Fatalf("third generation did not retain full history: %+v, err=%v",
			parsedThird, err)
	}
}

func TestRotateRejectsKeyReuseAndMissingRetiredCoverage(t *testing.T) {
	created := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	first, err := Build("018f0000-0000-7000-8000-000000000001", created,
		fixtureMaterial())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Rotate(first, created.Add(time.Hour), fixtureMaterial()); err == nil {
		t.Fatal("active key self-reuse was accepted")
	}
	second, err := Rotate(first, created.Add(time.Hour), rotatedMaterial(4))
	if err != nil {
		t.Fatal(err)
	}
	wire := second.wire
	wire.Retired = append([]keyWire(nil), wire.Retired[:2]...)
	raw, _ := json.Marshal(wire)
	tampered, _ := canonicaljsonv1.CanonicalizeObject(raw, MaximumBundleBytes)
	if _, err := Parse(tampered); err == nil {
		t.Fatal("rotation without retained coverage was accepted")
	}
}

func TestRotatedBundleResolvesHistoricalVerificationKey(t *testing.T) {
	created := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	first, err := Build("018f0000-0000-7000-8000-000000000001", created,
		fixtureMaterial())
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := Rotate(first, created.Add(time.Hour), rotatedMaterial(4))
	if err != nil {
		t.Fatal(err)
	}
	keyID, _ := first.KeyID(PurposeRecoveryManifest)
	expected, _ := first.PublicKey(PurposeRecoveryManifest)
	public, generation, found := rotated.VerificationKey(PurposeRecoveryManifest, keyID)
	if !found || generation != 1 || public != expected {
		t.Fatalf("historical key = found %v generation %d", found, generation)
	}
}

func rotatedMaterial(marker byte) []PublicMaterial {
	return []PublicMaterial{
		{Purpose: PurposeRecoveryManifest, PublicKey: [32]byte{marker}},
		{Purpose: PurposeRecoveryControl, PublicKey: [32]byte{marker + 1}},
		{Purpose: PurposeRestoreReceipt, PublicKey: [32]byte{marker + 2}},
	}
}
