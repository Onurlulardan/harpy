package backup

import (
	"context"
	_ "embed"
	"time"

	"github.com/google/uuid"
)

//go:embed sql/platform/get_receipt.sql
var getReceiptSQL string

type row interface{ Scan(...any) error }

type Receipt struct {
	ID, TenantID, PackageID, CatalogOperationID uuid.UUID
	OwnerGenerationsJSON                        []byte
	OwnerManifestRoot                           string
	DatabaseSnapshotRoot                        string
	ObjectInventoryRoot                         string
	KeyRevision                                 string
	CompletedAt                                 time.Time
	RestoreVerifiedAt                           time.Time
	ExpiresAt                                   time.Time
	VerifierVersion                             string
	LegalHold                                   bool
}

type Reader interface {
	Fetch(context.Context, uuid.UUID) (Receipt, error)
}

type reader struct {
	queryRow func(context.Context, string, ...any) row
	allowed  func(context.Context) bool
}

func (storage *reader) Fetch(ctx context.Context, receiptID uuid.UUID) (Receipt, error) {
	if storage == nil || storage.queryRow == nil || storage.allowed == nil ||
		ctx == nil || ctx.Err() != nil || !storage.allowed(ctx) || receiptID == uuid.Nil {
		return Receipt{}, ErrUnavailable
	}
	var value Receipt
	err := storage.queryRow(ctx, getReceiptSQL, receiptID).Scan(
		&value.ID, &value.TenantID, &value.PackageID, &value.CatalogOperationID,
		&value.OwnerGenerationsJSON, &value.OwnerManifestRoot, &value.DatabaseSnapshotRoot,
		&value.ObjectInventoryRoot, &value.KeyRevision, &value.CompletedAt,
		&value.RestoreVerifiedAt, &value.ExpiresAt, &value.VerifierVersion, &value.LegalHold,
	)
	if err != nil {
		return Receipt{}, err
	}
	value.CompletedAt = value.CompletedAt.UTC()
	value.RestoreVerifiedAt = value.RestoreVerifiedAt.UTC()
	value.ExpiresAt = value.ExpiresAt.UTC()
	return value, nil
}

var _ Reader = (*reader)(nil)
