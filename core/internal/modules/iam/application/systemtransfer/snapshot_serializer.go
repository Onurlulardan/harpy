package systemtransfer

import (
	"context"
	"errors"
	"io"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
)

func (worker *ExportWorker) serializeExportSnapshot(
	ctx context.Context,
	claim ExportJobClaim,
	attempt ExportAttempt,
	snapshot ExportSnapshot,
	destination io.Writer,
	progress *ExportProgress,
) error {
	format, err := tabulartransfer.ParseFormat(attempt.Format)
	if err != nil {
		return ErrExportQueryInvalid
	}
	bounded := &exportBoundedWriter{destination: destination,
		maximum: exportPlaintextMaximumBytes()}
	source := &exportCheckpointSource{ctx: ctx, worker: worker,
		claim: claim, attempt: attempt, source: snapshot.Rows,
		startedAt: snapshot.StartedAt, checkedAt: snapshot.StartedAt,
		bytes: bounded}
	result, err := tabulartransfer.Write(bounded, format,
		snapshot.Columns, source)
	now, nowErr := worker.exportNow()
	progress.Rows, progress.PlaintextBytes = source.rows, bounded.written
	progress.CheckedAt = now
	if nowErr != nil {
		return nowErr
	}
	if err != nil {
		if errors.Is(err, ErrExportLimit) {
			return err
		}
		return errors.Join(ErrExportStateConflict, err)
	}
	if result.Rows != source.rows || progress.Validate() != nil {
		return ErrExportStateConflict
	}
	return nil
}
