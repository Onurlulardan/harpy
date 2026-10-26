package backupreceipt

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type OwnerGeneration struct {
	Owner      string
	Generation int64
}

type Receipt struct {
	ID                   uuid.UUID
	TenantID             uuid.UUID
	PackageID            uuid.UUID
	CatalogOperationID   uuid.UUID
	OwnerGenerations     []OwnerGeneration
	OwnerManifestRoot    string
	DatabaseSnapshotRoot string
	ObjectInventoryRoot  string
	KeyRevision          string
	CompletedAt          time.Time
	RestoreVerifiedAt    time.Time
	ExpiresAt            time.Time
	VerifierVersion      string
	LegalHold            bool
}

type Requirement struct {
	TenantID           uuid.UUID
	PackageID          uuid.UUID
	CatalogOperationID uuid.UUID
	OwnerGenerations   []OwnerGeneration
	Now                time.Time
	MinimumExpiry      time.Time
}

func (value Receipt) ValidFor(requirement Requirement) error {
	if !value.valid() || !requirement.valid() {
		return ErrInvalid
	}
	if value.TenantID != requirement.TenantID || value.PackageID != requirement.PackageID ||
		value.CatalogOperationID != requirement.CatalogOperationID ||
		!sameGenerations(value.OwnerGenerations, requirement.OwnerGenerations) {
		return ErrBinding
	}
	if value.RestoreVerifiedAt.IsZero() || value.RestoreVerifiedAt.Before(requirement.Now.Add(-30*24*time.Hour)) ||
		value.RestoreVerifiedAt.After(requirement.Now) {
		return ErrNotVerified
	}
	if value.LegalHold {
		return ErrLegalHold
	}
	if value.ExpiresAt.Before(requirement.MinimumExpiry) {
		return ErrExpired
	}
	return nil
}

func (value Receipt) valid() bool {
	return value.ID != uuid.Nil && value.TenantID != uuid.Nil && value.PackageID != uuid.Nil &&
		value.CatalogOperationID != uuid.Nil && validGenerations(value.OwnerGenerations) &&
		digestPattern.MatchString(value.OwnerManifestRoot) &&
		digestPattern.MatchString(value.DatabaseSnapshotRoot) &&
		digestPattern.MatchString(value.ObjectInventoryRoot) &&
		value.KeyRevision != "" && len(value.KeyRevision) <= 128 &&
		canonicalTime(value.CompletedAt) && canonicalTime(value.RestoreVerifiedAt) &&
		canonicalTime(value.ExpiresAt) && value.ExpiresAt.After(value.RestoreVerifiedAt) &&
		value.VerifierVersion != "" && len(value.VerifierVersion) <= 64
}

func (value Requirement) valid() bool {
	return value.TenantID != uuid.Nil && value.PackageID != uuid.Nil &&
		value.CatalogOperationID != uuid.Nil && validGenerations(value.OwnerGenerations) &&
		canonicalTime(value.Now) && canonicalTime(value.MinimumExpiry) &&
		value.MinimumExpiry.After(value.Now)
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}
