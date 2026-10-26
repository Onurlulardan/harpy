package systemtransfer

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
)

func ReadSourceBatch(
	reader tabulartransfer.Reader,
	format tabulartransfer.Format,
	maximumRows int,
) ([]SourceRow, bool, error) {
	if reader == nil || !format.Importable() || maximumRows < 1 ||
		maximumRows > tabulartransfer.MaximumPageRows {
		return nil, false, ErrInvalidAttempt
	}
	rows := make([]SourceRow, 0, maximumRows)
	for len(rows) < maximumRows {
		row, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return rows, true, nil
		}
		if err != nil {
			return nil, false, err
		}
		converted, convertErr := sourceRowFromTabular(format, row)
		if convertErr != nil {
			return nil, false, convertErr
		}
		rows = append(rows, converted)
	}
	return rows, false, nil
}

func sourceRowFromTabular(
	format tabulartransfer.Format,
	row tabulartransfer.Row,
) (SourceRow, error) {
	if row.Index < 1 || len(row.Values) < 1 ||
		len(row.Values) > tabulartransfer.MaximumColumns {
		return SourceRow{}, ErrInvalidAttempt
	}
	values := make(map[string]json.RawMessage, len(row.Values))
	for key, value := range row.Values {
		if key == "" {
			return SourceRow{}, ErrInvalidAttempt
		}
		if format == tabulartransfer.FormatCSV {
			encoded, _ := json.Marshal(value)
			values[key] = encoded
			continue
		}
		if !json.Valid([]byte(value)) {
			return SourceRow{}, ErrInvalidAttempt
		}
		values[key] = json.RawMessage(value)
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return SourceRow{}, ErrInvalidAttempt
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		encoded, int(tabulartransfer.MaximumRowBytes),
	)
	if err != nil {
		return SourceRow{}, ErrInvalidAttempt
	}
	return SourceRow{Index: row.Index, JSON: canonical,
		SHA256: sha256.Sum256(canonical)}, nil
}
