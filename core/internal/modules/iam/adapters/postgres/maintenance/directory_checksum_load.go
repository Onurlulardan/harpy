package maintenance

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

func (store *store) loadDirectoryChecksumSubjects(
	ctx context.Context,
	tenantID uuid.UUID,
	claimValues []any,
) ([]directoryChecksumSubject, error) {
	subjects := make(map[uuid.UUID]*directoryChecksumSubject)
	wipe := func() {
		for _, subject := range subjects {
			subject.wipe()
		}
	}
	rows, err := store.database.Query(ctx, readDirectorySourcesSQL, claimValues...)
	if err != nil {
		return nil, fmt.Errorf("read directory sources: %w", err)
	}
	for rows.Next() {
		source, scanErr := scanDirectorySource(rows)
		if scanErr != nil {
			rows.Close()
			wipe()
			return nil, scanErr
		}
		if _, duplicate := subjects[source.id]; duplicate {
			source.wipe()
			rows.Close()
			wipe()
			return nil, ErrRuntimeInvalid
		}
		copy := source
		subjects[source.id] = &directoryChecksumSubject{id: source.id, source: &copy}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		wipe()
		return nil, fmt.Errorf("iterate directory sources: %w", err)
	}
	rows.Close()
	rows, err = store.database.Query(ctx, readDirectorySnapshotSQL, tenantID)
	if err != nil {
		wipe()
		return nil, fmt.Errorf("read directory projections: %w", err)
	}
	for rows.Next() {
		sourceID, actual, scanErr := scanDirectoryActual(rows)
		if scanErr != nil {
			rows.Close()
			wipe()
			return nil, scanErr
		}
		subject := subjects[sourceID]
		if subject == nil {
			subject = &directoryChecksumSubject{id: sourceID}
			subjects[sourceID] = subject
		}
		subject.actuals = append(subject.actuals, actual)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		wipe()
		return nil, fmt.Errorf("iterate directory projections: %w", err)
	}
	rows.Close()
	ids := make([]uuid.UUID, 0, len(subjects))
	for id := range subjects {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left, right int) bool {
		return bytes.Compare(ids[left][:], ids[right][:]) < 0
	})
	result := make([]directoryChecksumSubject, 0, len(ids))
	for _, id := range ids {
		result = append(result, *subjects[id])
	}
	return result, nil
}
