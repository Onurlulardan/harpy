package demomaterialization

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type nullableTime struct {
	value time.Time
	valid bool
}

func (value *nullableTime) Scan(source any) error {
	if source == nil {
		value.value = time.Time{}
		value.valid = false
		return nil
	}
	scanned, ok := source.(time.Time)
	if !ok {
		return ErrUnavailable
	}
	value.value = scanned
	value.valid = true
	return nil
}

func scanSnapshot(source Row) (Snapshot, error) {
	var snapshot Snapshot
	var componentID uuid.UUID
	var semanticSHA256, requestSHA256 []byte
	var retryAt nullableTime
	err := source.Scan(
		&snapshot.ReceiptID, &snapshot.PackageID, &snapshot.PackageSlug,
		&snapshot.PackageVersion, &snapshot.ReleaseID, &componentID,
		&snapshot.Component.Generation, &snapshot.Component.Contract,
		&snapshot.Component.ContentSHA256, &snapshot.Materialization.ModuleSlug,
		&snapshot.Materialization.SourceContract,
		&semanticSHA256, &snapshot.Materialization.FormCount,
		&snapshot.Materialization.RowCount, &snapshot.Materialization.Document,
		&snapshot.IdempotencyKey, &requestSHA256, &snapshot.State,
		&snapshot.Ready, &retryAt, &snapshot.CompletedAt,
		&snapshot.CreatedAt, &snapshot.UpdatedAt,
	)
	if err != nil {
		var stateError interface{ SQLState() string }
		if errors.As(err, &stateError) && stateError.SQLState() == "23505" {
			return Snapshot{}, ErrConflict
		}
		return Snapshot{}, fmt.Errorf(
			"scan Catalog demo materialization: %w", ErrUnavailable,
		)
	}
	if len(semanticSHA256) != 32 || len(requestSHA256) != 32 {
		return Snapshot{}, fmt.Errorf("scan Catalog demo materialization: %w", ErrUnavailable)
	}
	snapshot.Component.ID = componentID.String()
	snapshot.Component.Owner = owner
	copy(snapshot.Materialization.SemanticSHA256[:], semanticSHA256)
	copy(snapshot.RequestSHA256[:], requestSHA256)
	if retryAt.valid {
		snapshot.RetryAt = retryAt.value.UTC()
	}
	snapshot.CompletedAt = snapshot.CompletedAt.UTC()
	snapshot.CreatedAt = snapshot.CreatedAt.UTC()
	snapshot.UpdatedAt = snapshot.UpdatedAt.UTC()
	return snapshot, nil
}
