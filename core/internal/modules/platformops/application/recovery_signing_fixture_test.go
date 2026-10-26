package application

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type recoveryStoreFixture struct {
	records map[int64]ports.SigningBundleRecord
}

func (store *recoveryStoreFixture) RecordSigningBundle(
	_ context.Context,
	record ports.SigningBundleRecord,
) error {
	if current, found := store.records[record.Generation]; found {
		if current.SHA256 == record.SHA256 {
			return nil
		}
		return ports.ErrConflict
	}
	if record.Generation > 1 {
		if _, found := store.records[record.Generation-1]; !found {
			return ports.ErrConflict
		}
	}
	store.records[record.Generation] = record
	return nil
}

func (store *recoveryStoreFixture) LoadSigningBundle(
	_ context.Context,
	_ string,
	generation int64,
) (ports.SigningBundleRecord, error) {
	record, found := store.records[generation]
	if !found {
		return ports.SigningBundleRecord{}, ports.ErrNotFound
	}
	return record, nil
}

type recoveryClockFixture struct{ now time.Time }

func (clock recoveryClockFixture) Now() time.Time { return clock.now }

func buildRecoveryBundleFixture(t *testing.T, now time.Time) recoverysigningv1.Bundle {
	t.Helper()
	purposes := []recoverysigningv1.Purpose{recoverysigningv1.PurposeRecoveryManifest,
		recoverysigningv1.PurposeRecoveryControl, recoverysigningv1.PurposeRestoreReceipt}
	materials := make([]recoverysigningv1.PublicMaterial, 0, len(purposes))
	for _, purpose := range purposes {
		public, _, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		var encoded [32]byte
		copy(encoded[:], public)
		materials = append(materials, recoverysigningv1.PublicMaterial{Purpose: purpose, PublicKey: encoded})
	}
	bundle, err := recoverysigningv1.Build("018f0000-0000-7000-8000-000000000001", now, materials)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}
