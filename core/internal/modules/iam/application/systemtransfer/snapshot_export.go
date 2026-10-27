package systemtransfer

import (
	"context"
	"time"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
	"github.com/google/uuid"
)

const MaximumExportSnapshotAge = 30 * time.Minute

type ExportSnapshot struct {
	CatalogVersionID           uuid.UUID
	CatalogSHA256, QuerySHA256 [32]byte
	StartedAt                  time.Time
	Columns                    []string
	Rows                       tabulartransfer.RowSource
}

func (snapshot ExportSnapshot) Validate(attempt ExportAttempt) error {
	if snapshot.CatalogVersionID != attempt.CatalogVersionID ||
		snapshot.CatalogSHA256 != attempt.CatalogSHA256 ||
		snapshot.QuerySHA256 != attempt.QuerySHA256 ||
		!exportArtifactTime(snapshot.StartedAt) || len(snapshot.Columns) < 1 ||
		len(snapshot.Columns) > tabulartransfer.MaximumColumns ||
		nilDependency(snapshot.Rows) {
		return ErrExportCatalogPinStale
	}
	seen := make(map[string]struct{}, len(snapshot.Columns))
	for _, column := range snapshot.Columns {
		if column == "" {
			return ErrExportQueryInvalid
		}
		if _, exists := seen[column]; exists {
			return ErrExportQueryInvalid
		}
		seen[column] = struct{}{}
	}
	return nil
}

// ExportSnapshotProvider opens the exact retained catalog version and digest
// pinned by the owner row. Active-catalog substitution is forbidden. It must
// consume the callback synchronously and return only after both the row source
// and its database snapshot are closed.
type ExportSnapshotProvider interface {
	WithPinnedSnapshot(
		context.Context,
		ExportAttempt,
		func(context.Context, ExportSnapshot) error,
	) error
}
