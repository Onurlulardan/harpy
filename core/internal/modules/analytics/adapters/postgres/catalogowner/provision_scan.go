package catalogowner

import (
	"fmt"
	"time"

	catalogprojection "github.com/Onurlulardan/harpy/core/internal/capabilities/catalogprojection/api"
	"github.com/google/uuid"
)

func scanProvisionSnapshot(
	source catalogprojection.Row,
) (catalogprojection.Snapshot, error) {
	var snapshot catalogprojection.Snapshot
	var componentID uuid.UUID
	var semanticSHA256, requestSHA256 []byte
	var readyAt *time.Time
	err := source.Scan(
		&snapshot.ReceiptID, &snapshot.PackageID, &snapshot.PackageSlug,
		&snapshot.PackageVersion, &snapshot.ReleaseID, &componentID,
		&snapshot.Component.Generation, &snapshot.Component.Contract,
		&snapshot.Component.ContentSHA256, &snapshot.Materialization.Kind,
		&snapshot.Materialization.ModuleSlug, &snapshot.Materialization.Identity,
		&semanticSHA256, &snapshot.Materialization.ItemCount,
		&snapshot.Materialization.AuxiliaryCount,
		&snapshot.Materialization.Document, &snapshot.IdempotencyKey,
		&requestSHA256, &snapshot.State, &readyAt,
		&snapshot.CreatedAt, &snapshot.UpdatedAt,
	)
	if err != nil || len(semanticSHA256) != 32 || len(requestSHA256) != 32 {
		return catalogprojection.Snapshot{}, fmt.Errorf(
			"scan Analytics catalog materialization: %w",
			catalogprojection.ErrUnavailable,
		)
	}
	snapshot.Component.ID, snapshot.Component.Owner = componentID.String(), owner
	copy(snapshot.Materialization.SemanticSHA256[:], semanticSHA256)
	copy(snapshot.RequestSHA256[:], requestSHA256)
	if readyAt != nil {
		snapshot.ReadyAt = readyAt.UTC()
	}
	snapshot.CreatedAt = snapshot.CreatedAt.UTC()
	snapshot.UpdatedAt = snapshot.UpdatedAt.UTC()
	return snapshot, nil
}
