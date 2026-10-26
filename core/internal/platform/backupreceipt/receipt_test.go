package backupreceipt

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestReceiptRejectsBindingAndOwnerGenerationTampering(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	receipt := receiptFixture(now)
	requirement := Requirement{TenantID: receipt.TenantID, PackageID: receipt.PackageID,
		CatalogOperationID: receipt.CatalogOperationID,
		OwnerGenerations:   []OwnerGeneration{{Owner: "forms", Generation: 7}},
		Now:                now, MinimumExpiry: now.Add(90 * 24 * time.Hour)}
	if err := receipt.ValidFor(requirement); err != nil {
		t.Fatalf("valid receipt rejected: %v", err)
	}
	tampered := requirement
	tampered.CatalogOperationID = uuid.MustParse("018f3000-0000-7000-8000-000000000099")
	if err := receipt.ValidFor(tampered); !errors.Is(err, ErrBinding) {
		t.Fatalf("operation tamper error = %v", err)
	}
	tampered = requirement
	tampered.OwnerGenerations = []OwnerGeneration{{Owner: "forms", Generation: 8}}
	if err := receipt.ValidFor(tampered); !errors.Is(err, ErrBinding) {
		t.Fatalf("generation tamper error = %v", err)
	}
}

func TestReceiptRejectsUnverifiedOrMalformedEvidence(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	receipt := receiptFixture(now)
	requirement := Requirement{TenantID: receipt.TenantID, PackageID: receipt.PackageID,
		CatalogOperationID: receipt.CatalogOperationID, OwnerGenerations: receipt.OwnerGenerations,
		Now: now, MinimumExpiry: now.Add(90 * 24 * time.Hour)}
	receipt.RestoreVerifiedAt = time.Time{}
	if err := receipt.ValidFor(requirement); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unverified receipt error = %v", err)
	}
	receipt = receiptFixture(now)
	receipt.OwnerManifestRoot = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	if err := receipt.ValidFor(requirement); !errors.Is(err, ErrInvalid) {
		t.Fatalf("malformed digest error = %v", err)
	}
}

func TestReceiptRejectsLegalHold(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	receipt := receiptFixture(now)
	receipt.LegalHold = true
	requirement := Requirement{TenantID: receipt.TenantID, PackageID: receipt.PackageID,
		CatalogOperationID: receipt.CatalogOperationID, OwnerGenerations: receipt.OwnerGenerations,
		Now: now, MinimumExpiry: now.Add(90 * 24 * time.Hour)}
	if err := receipt.ValidFor(requirement); !errors.Is(err, ErrLegalHold) {
		t.Fatalf("legal hold error = %v", err)
	}
}

func receiptFixture(now time.Time) Receipt {
	return Receipt{
		ID:                   uuid.MustParse("018f3000-0000-7000-8000-000000000001"),
		TenantID:             uuid.MustParse("018f3000-0000-7000-8000-000000000002"),
		PackageID:            uuid.MustParse("018f3000-0000-7000-8000-000000000003"),
		CatalogOperationID:   uuid.MustParse("018f3000-0000-7000-8000-000000000004"),
		OwnerGenerations:     []OwnerGeneration{{Owner: "forms", Generation: 7}},
		OwnerManifestRoot:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		DatabaseSnapshotRoot: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ObjectInventoryRoot:  "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		KeyRevision:          "kms-v7", CompletedAt: now.Add(-48 * time.Hour),
		RestoreVerifiedAt: now.Add(-24 * time.Hour), ExpiresAt: now.Add(180 * 24 * time.Hour),
		VerifierVersion: "restore-verifier.v1",
	}
}
