package dashboardexport

import (
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

type snapshotRow struct {
	exportID, dashboardID, dashboardVersionID, jobID uuid.UUID
	effectiveState, storedState, format              string
	filename, contentType, failureCode               string
	rowCount, plaintextBytes                         int64
	cancellationRequested                            bool
	createdAt, updatedAt, expiresAt                  time.Time
}

func scanSnapshot(scanner row, notFound rowNotFound) (exportapp.Snapshot, error) {
	stored := snapshotRow{}
	err := scanner.Scan(&stored.exportID, &stored.dashboardID, &stored.dashboardVersionID,
		&stored.jobID, &stored.effectiveState, &stored.storedState, &stored.format,
		&stored.filename, &stored.contentType, &stored.failureCode, &stored.rowCount,
		&stored.plaintextBytes, &stored.cancellationRequested, &stored.createdAt,
		&stored.updatedAt, &stored.expiresAt)
	if err != nil {
		return exportapp.Snapshot{}, mapError(err, notFound)
	}
	return stored.snapshot(), nil
}

func scanSubmit(scanner row, notFound rowNotFound) (exportapp.Snapshot, bool, error) {
	stored := snapshotRow{}
	var replay bool
	err := scanner.Scan(&stored.exportID, &stored.dashboardID, &stored.dashboardVersionID,
		&stored.jobID, &stored.effectiveState, &stored.storedState, &stored.format,
		&stored.filename, &stored.contentType, &stored.failureCode, &stored.rowCount,
		&stored.plaintextBytes, &stored.cancellationRequested, &stored.createdAt,
		&stored.updatedAt, &stored.expiresAt, &replay)
	if err != nil {
		return exportapp.Snapshot{}, false, mapError(err, notFound)
	}
	return stored.snapshot(), replay, nil
}

func scanDelete(scanner row, notFound rowNotFound) (
	exportapp.Snapshot, bool, bool, error,
) {
	stored := snapshotRow{}
	var replay, schedule bool
	err := scanner.Scan(&stored.exportID, &stored.dashboardID, &stored.dashboardVersionID,
		&stored.jobID, &stored.effectiveState, &stored.storedState, &stored.format,
		&stored.filename, &stored.contentType, &stored.failureCode, &stored.rowCount,
		&stored.plaintextBytes, &stored.cancellationRequested, &stored.createdAt,
		&stored.updatedAt, &stored.expiresAt, &replay, &schedule)
	if err != nil {
		return exportapp.Snapshot{}, false, false, mapError(err, notFound)
	}
	return stored.snapshot(), replay, schedule, nil
}

func (stored snapshotRow) snapshot() exportapp.Snapshot {
	return exportapp.Snapshot{ExportID: stored.exportID, DashboardID: stored.dashboardID,
		DashboardVersionID: stored.dashboardVersionID, JobID: stored.jobID,
		State: exportapp.State(stored.effectiveState), StoredState: exportapp.State(stored.storedState),
		Format: exportapp.Format(stored.format), Filename: stored.filename,
		ContentType: stored.contentType, FailureCode: stored.failureCode,
		RowCount: stored.rowCount, PlaintextBytes: stored.plaintextBytes,
		CancellationRequested: stored.cancellationRequested,
		CreatedAt:             stored.createdAt.UTC(), UpdatedAt: stored.updatedAt.UTC(),
		ExpiresAt: stored.expiresAt.UTC()}
}
