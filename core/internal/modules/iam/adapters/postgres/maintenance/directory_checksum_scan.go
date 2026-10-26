package maintenance

import (
	"fmt"

	"github.com/google/uuid"
)

func scanDirectorySource(rows rowScanner) (directorySourceRow, error) {
	var source directorySourceRow
	if err := rows.Scan(
		&source.id, &source.userID, &source.kind, &source.normalized,
		&source.normalizedSHA, &source.state, &source.sourceRevision,
	); err != nil {
		return directorySourceRow{}, fmt.Errorf("scan directory source: %w", err)
	}
	if !source.valid() {
		source.wipe()
		return directorySourceRow{}, ErrRuntimeInvalid
	}
	return source, nil
}

func scanDirectoryActual(rows rowScanner) (uuid.UUID, directoryActualRow, error) {
	var sourceID uuid.UUID
	var actual directoryActualRow
	if err := rows.Scan(
		&sourceID, &actual.keyID, &actual.hmac, &actual.userID,
		&actual.sourceRevision, &actual.kind, &actual.state, &actual.sourceSHA,
	); err != nil {
		return uuid.Nil, directoryActualRow{}, fmt.Errorf("scan directory projection: %w", err)
	}
	if sourceID == uuid.Nil || !actual.valid() {
		clear(actual.hmac)
		clear(actual.sourceSHA)
		return uuid.Nil, directoryActualRow{}, ErrRuntimeInvalid
	}
	return sourceID, actual, nil
}
