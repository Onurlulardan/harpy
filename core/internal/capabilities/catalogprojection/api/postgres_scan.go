package catalogprojection

import (
	"fmt"

	"github.com/google/uuid"
)

func scanSnapshot(source Row) (Snapshot, error) {
	var snapshot Snapshot
	var componentID uuid.UUID
	var semanticSHA256, requestSHA256 []byte
	err := source.Scan(
		&snapshot.ReceiptID, &snapshot.PackageID, &snapshot.PackageSlug,
		&snapshot.PackageVersion,
		&snapshot.ReleaseID, &componentID, &snapshot.Component.Generation,
		&snapshot.Component.Contract, &snapshot.Component.ContentSHA256,
		&snapshot.Materialization.Kind, &snapshot.Materialization.ModuleSlug,
		&snapshot.Materialization.Identity,
		&semanticSHA256, &snapshot.Materialization.ItemCount,
		&snapshot.Materialization.AuxiliaryCount, &snapshot.Materialization.Document,
		&snapshot.IdempotencyKey, &requestSHA256, &snapshot.State,
		&snapshot.ReadyAt, &snapshot.CreatedAt, &snapshot.UpdatedAt,
	)
	if err != nil || len(semanticSHA256) != 32 || len(requestSHA256) != 32 {
		return Snapshot{}, fmt.Errorf("scan catalog materialization: %w", ErrUnavailable)
	}
	snapshot.Component.ID = componentID.String()
	copy(snapshot.Materialization.SemanticSHA256[:], semanticSHA256)
	copy(snapshot.RequestSHA256[:], requestSHA256)
	snapshot.ReadyAt = snapshot.ReadyAt.UTC()
	snapshot.CreatedAt = snapshot.CreatedAt.UTC()
	snapshot.UpdatedAt = snapshot.UpdatedAt.UTC()
	return snapshot, nil
}
