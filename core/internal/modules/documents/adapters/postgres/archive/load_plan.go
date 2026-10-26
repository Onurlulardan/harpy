package archive

import (
	"context"
	"time"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Pin(ctx context.Context, id domain.ID,
	now time.Time,
) (archiveapp.Plan, error) {
	if !storage.usable(ctx) || !storage.job || !storage.command || !id.UUIDv7() ||
		now.IsZero() {
		return archiveapp.Plan{}, archiveapp.ErrInvalid
	}
	if _, err := storage.execute(ctx, pinSQL, id.String(), now.UTC()); err != nil {
		return archiveapp.Plan{}, archiveapp.ErrUnavailable
	}
	snapshot, err := storage.Get(ctx, id)
	if err != nil {
		return archiveapp.Plan{}, err
	}
	entries, err := storage.listEntries(ctx, id)
	if err != nil {
		return archiveapp.Plan{}, err
	}
	if snapshot.State != domain.ArchivePending || snapshot.EntryCount != len(entries) ||
		len(entries) == 0 || snapshot.EntriesSHA256 == [32]byte{} {
		return archiveapp.Plan{}, archiveapp.ErrLimit
	}
	return archiveapp.Plan{Snapshot: snapshot, Entries: entries}, nil
}

func (storage *store) LoadPlan(ctx context.Context,
	id domain.ID,
) (archiveapp.Plan, error) {
	if !storage.usable(ctx) || !id.UUIDv7() {
		return archiveapp.Plan{}, archiveapp.ErrInvalid
	}
	snapshot, err := storage.Get(ctx, id)
	if err != nil {
		return archiveapp.Plan{}, err
	}
	entries, err := storage.listEntries(ctx, id)
	if err != nil {
		return archiveapp.Plan{}, err
	}
	return archiveapp.Plan{Snapshot: snapshot, Entries: entries}, nil
}

func (storage *store) listEntries(ctx context.Context,
	id domain.ID,
) ([]archiveapp.Entry, error) {
	values, err := storage.query(ctx, entriesSQL, id.String())
	if err != nil {
		return nil, archiveapp.ErrUnavailable
	}
	defer values.Close()
	entries := make([]archiveapp.Entry, 0)
	for values.Next() {
		var documentID, versionID string
		var entry archiveapp.Entry
		var digest, cipherDigest, wrapped []byte
		var generation int64
		if err = values.Scan(&documentID, &versionID, &entry.Order,
			&entry.PlaintextBytes, &digest, &entry.OriginalName,
			&entry.Object.Key, &entry.Object.Version, &generation,
			&entry.Object.CiphertextBytes, &cipherDigest, &entry.Object.KeyID,
			&wrapped); err != nil {
			return nil, archiveapp.ErrUnavailable
		}
		entry.TenantID = storage.tenantID
		entry.DocumentID, err = domain.ParseID(documentID)
		if err == nil {
			entry.VersionID, err = domain.ParseID(versionID)
		}
		if err != nil || len(digest) != 32 || len(cipherDigest) != 32 ||
			generation < 1 || len(wrapped) < 115 {
			return nil, archiveapp.ErrUnavailable
		}
		copy(entry.PlaintextSHA256[:], digest)
		entry.Object.Generation = uint64(generation)
		entry.Object.PlaintextBytes = entry.PlaintextBytes
		entry.Object.PlaintextSHA256 = entry.PlaintextSHA256
		copy(entry.Object.CiphertextSHA256[:], cipherDigest)
		entry.Object.WrappedDEK = append([]byte(nil), wrapped...)
		entries = append(entries, entry)
	}
	if values.Err() != nil {
		return nil, archiveapp.ErrUnavailable
	}
	return entries, nil
}
