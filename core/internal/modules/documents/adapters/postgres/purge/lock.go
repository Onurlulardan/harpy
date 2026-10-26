package purge

import (
	"context"
	"strings"
	"time"

	purgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/purge"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) Lock(ctx context.Context, itemID domain.ID,
	now time.Time,
) (purgeapp.Candidate, error) {
	if !storage.usable(ctx) || !itemID.UUIDv7() || now.IsZero() {
		return purgeapp.Candidate{}, purgeapp.ErrInvalid
	}
	result := purgeapp.Candidate{ItemID: itemID, TenantID: storage.tenantID}
	values, err := storage.query(ctx, lockSQL, itemID.String(), now.UTC())
	if err != nil {
		return purgeapp.Candidate{}, purgeapp.ErrUnavailable
	}
	defer values.Close()
	var seen, eligible, held, referenced bool
	for values.Next() {
		var id, owner, state string
		var versionID, key, version dbText
		var generation, plaintext, ciphertext dbInt64
		if err = values.Scan(&id, &owner, &state, &eligible, &held, &referenced,
			&versionID, &key, &version, &generation, &plaintext, &ciphertext); err != nil {
			return purgeapp.Candidate{}, purgeapp.ErrUnavailable
		}
		if !seen {
			result.OwnerID, err = domain.ParseID(owner)
			if err != nil || id != itemID.String() || !purgeState(state, &result.State) {
				return purgeapp.Candidate{}, purgeapp.ErrUnavailable
			}
			seen = true
		}
		if versionID.Valid {
			parsed, parseErr := domain.ParseID(versionID.Text)
			if parseErr != nil || !key.Valid || !version.Valid || generation.Int64 < 1 {
				return purgeapp.Candidate{}, purgeapp.ErrUnavailable
			}
			result.Versions = append(result.Versions, purgeapp.Version{ID: parsed,
				ObjectKey: key.Text, ObjectVersion: version.Text,
				ObjectGeneration: uint64(generation.Int64),
				PlaintextBytes:   plaintext.Int64, CiphertextBytes: ciphertext.Int64})
		}
	}
	if values.Err() != nil {
		return purgeapp.Candidate{}, purgeapp.ErrUnavailable
	}
	if !seen {
		return purgeapp.Candidate{}, purgeapp.ErrNotFound
	}
	if held {
		return purgeapp.Candidate{}, purgeapp.ErrLegalHold
	}
	if referenced {
		return purgeapp.Candidate{}, purgeapp.ErrReference
	}
	if !eligible {
		return purgeapp.Candidate{}, purgeapp.ErrConflict
	}
	return result, nil
}

func purgeState(value string, target *domain.ItemState) bool {
	switch strings.ToUpper(value) {
	case "TRASHED":
		*target = domain.ItemTrashed
	case "PURGE_COOLING":
		*target = domain.ItemPurgeCooling
	case "PURGING":
		*target = domain.ItemPurging
	default:
		return false
	}
	return true
}
