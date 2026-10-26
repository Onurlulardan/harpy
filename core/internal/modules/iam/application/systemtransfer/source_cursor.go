package systemtransfer

import tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"

type SourceCursor struct {
	reader  tabulartransfer.Reader
	format  tabulartransfer.Format
	pending *SourceRow
	next    int64
}

func NewSourceCursor(reader tabulartransfer.Reader,
	format tabulartransfer.Format) (*SourceCursor, error) {
	if reader == nil || !format.Importable() {
		return nil, ErrInvalidAttempt
	}
	return &SourceCursor{reader: reader, format: format, next: 1}, nil
}

func (cursor *SourceCursor) SkipProcessed(processed int64) error {
	if cursor == nil || processed < 0 || processed > 1_000_000 ||
		cursor.next != 1 || cursor.pending != nil {
		return ErrInvalidAttempt
	}
	for index := int64(1); index <= processed; index++ {
		row, eof, err := cursor.readOne()
		if err != nil || eof || row.Index != index {
			return ErrWorkerState
		}
		cursor.next++
	}
	return nil
}

func (cursor *SourceCursor) NextBatch() ([]SourceRow, bool, error) {
	if cursor == nil || cursor.reader == nil || cursor.next < 1 {
		return nil, false, ErrInvalidAttempt
	}
	rows := make([]SourceRow, 0, MaximumBatchRows)
	var bytes int64
	for len(rows) < MaximumBatchRows {
		row, eof, err := cursor.nextOne()
		if err != nil {
			return nil, false, err
		}
		if eof {
			return rows, true, nil
		}
		if row.Index != cursor.next {
			return nil, false, ErrWorkerState
		}
		size := int64(len(row.JSON))
		if len(rows) > 0 && bytes+size > MaximumMappedBatchByte {
			copy := row
			cursor.pending = &copy
			return rows, false, nil
		}
		if size > MaximumMappedBatchByte {
			return nil, false, ErrInvalidAttempt
		}
		rows = append(rows, row)
		bytes += size
		cursor.next++
	}
	return rows, false, nil
}

func (cursor *SourceCursor) nextOne() (SourceRow, bool, error) {
	if cursor.pending != nil {
		row := *cursor.pending
		cursor.pending = nil
		return row, false, nil
	}
	return cursor.readOne()
}

func (cursor *SourceCursor) readOne() (SourceRow, bool, error) {
	rows, eof, err := ReadSourceBatch(cursor.reader, cursor.format, 1)
	if err != nil {
		return SourceRow{}, false, err
	}
	if len(rows) == 0 {
		if !eof {
			return SourceRow{}, false, ErrWorkerState
		}
		return SourceRow{}, true, nil
	}
	return rows[0], false, nil
}
