package systemtransfer

import (
	"context"
	"io"
	"time"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
)

const exportCheckpointAge = 20 * time.Second

type exportCheckpointSource struct {
	ctx       context.Context
	worker    *ExportWorker
	claim     ExportJobClaim
	attempt   ExportAttempt
	source    tabulartransfer.RowSource
	startedAt time.Time
	checkedAt time.Time
	bytes     *exportBoundedWriter
	rows      int64
}

func (source *exportCheckpointSource) Next() ([]tabulartransfer.Cell, error) {
	now, err := source.worker.exportNow()
	if err != nil {
		return nil, err
	}
	if now.Before(source.startedAt) || now.Sub(source.startedAt) > MaximumExportSnapshotAge {
		return nil, ErrExportLimit
	}
	if source.rows > 0 && (source.rows%tabulartransfer.MaximumPageRows == 0 ||
		now.Sub(source.checkedAt) >= exportCheckpointAge) {
		progress := ExportProgress{Rows: source.rows,
			PlaintextBytes: source.bytes.written, CheckedAt: now}
		if err = source.worker.checkpointExport(
			source.ctx, source.claim, source.attempt, progress,
		); err != nil {
			return nil, err
		}
		source.checkedAt = now
	}
	cells, err := source.source.Next()
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}
	if source.rows >= tabulartransfer.MaximumRows || exportRowBytes(cells) >
		tabulartransfer.MaximumRowBytes {
		return nil, ErrExportLimit
	}
	source.rows++
	return cells, nil
}

func exportRowBytes(cells []tabulartransfer.Cell) int64 {
	var size int64
	for _, cell := range cells {
		size += int64(len(cell.Text))
	}
	return size
}
