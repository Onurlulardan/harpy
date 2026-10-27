package purge

import (
	"context"

	purgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/purge"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) ScanPending(ctx context.Context, itemID,
	afterDeletionID domain.ID, limit uint16,
) (purgeapp.DeletionPage, error) {
	if !storage.usable(ctx) || !itemID.UUIDv7() ||
		afterDeletionID.Valid() && !afterDeletionID.UUIDv7() || limit == 0 || limit > 10_000 {
		return purgeapp.DeletionPage{}, purgeapp.ErrInvalid
	}
	var after any
	if afterDeletionID.Valid() {
		after = afterDeletionID.String()
	}
	rows, err := storage.query(ctx, listPendingSQL, itemID.String(), after, int(limit)+1)
	if err != nil {
		return purgeapp.DeletionPage{}, purgeapp.ErrUnavailable
	}
	defer rows.Close()
	page := purgeapp.DeletionPage{Items: make([]purgeapp.Deletion, 0, limit)}
	for rows.Next() {
		var values [2]string
		var key, version string
		var generation, plaintext, ciphertext int64
		if err = rows.Scan(&values[0], &values[1], &key, &version,
			&generation, &plaintext, &ciphertext); err != nil {
			return purgeapp.DeletionPage{}, purgeapp.ErrUnavailable
		}
		deletionID, deletionErr := domain.ParseID(values[0])
		versionID, versionErr := domain.ParseID(values[1])
		if deletionErr != nil || versionErr != nil || generation < 1 ||
			key == "" || version == "" || plaintext < 0 || ciphertext < 1 {
			return purgeapp.DeletionPage{}, purgeapp.ErrUnavailable
		}
		if len(page.Items) == int(limit) {
			page.More = true
			continue
		}
		page.Items = append(page.Items, purgeapp.Deletion{
			Version: purgeapp.Version{ID: versionID, ObjectKey: key,
				ObjectVersion: version, ObjectGeneration: uint64(generation),
				PlaintextBytes: plaintext, CiphertextBytes: ciphertext},
			DeletionID: deletionID, JobID: deletionID,
		})
	}
	if err = rows.Err(); err != nil {
		return purgeapp.DeletionPage{}, purgeapp.ErrUnavailable
	}
	return page, nil
}
