package item

import (
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type versionDocument struct {
	ID              string    `json:"id"`
	Number          uint64    `json:"number"`
	OriginalName    string    `json:"original_name"`
	MIME            string    `json:"mime"`
	FileKind        string    `json:"file_kind"`
	PlaintextBytes  int64     `json:"plaintext_bytes"`
	PlaintextSHA256 string    `json:"plaintext_sha256"`
	CreatedAt       time.Time `json:"created_at"`
}

type itemDocument struct {
	ID               string          `json:"id"`
	FolderID         string          `json:"folder_id"`
	OwnerID          string          `json:"owner_id"`
	CurrentVersionID string          `json:"current_version_id"`
	State            string          `json:"state"`
	RowVersion       uint64          `json:"row_version"`
	Description      string          `json:"description"`
	Tags             []string        `json:"tags"`
	Current          versionDocument `json:"current"`
	TrashedAt        *time.Time      `json:"trashed_at"`
	RetentionEndsAt  *time.Time      `json:"retention_ends_at"`
	PurgeRequestedAt *time.Time      `json:"purge_requested_at"`
	PurgeStartsAt    *time.Time      `json:"purge_starts_at"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func decodeItem(tenantID domain.ID, document []byte) (itemapp.Snapshot, error) {
	var wire itemDocument
	if json.Unmarshal(document, &wire) != nil {
		return itemapp.Snapshot{}, itemapp.ErrUnavailable
	}
	id, idErr := domain.ParseID(wire.ID)
	folderID, folderErr := domain.ParseID(wire.FolderID)
	ownerID, ownerErr := domain.ParseID(wire.OwnerID)
	currentID, currentErr := domain.ParseID(wire.CurrentVersionID)
	versionID, versionErr := domain.ParseID(wire.Current.ID)
	state, stateOK := decodeState(wire.State)
	digest, digestErr := hex.DecodeString(wire.Current.PlaintextSHA256)
	description, tags, _, metadataOK := foundation.NormalizeItemMetadata(
		wire.Description, wire.Tags)
	if idErr != nil || folderErr != nil || ownerErr != nil || currentErr != nil ||
		versionErr != nil || !stateOK || wire.RowVersion == 0 || !metadataOK ||
		description != wire.Description || !slices.Equal(tags, wire.Tags) ||
		wire.Current.Number == 0 || wire.Current.OriginalName == "" ||
		wire.Current.MIME == "" || wire.Current.FileKind == "" ||
		wire.Current.PlaintextBytes < 1 || digestErr != nil || len(digest) != 32 ||
		wire.CreatedAt.IsZero() || wire.UpdatedAt.IsZero() || wire.Current.CreatedAt.IsZero() {
		return itemapp.Snapshot{}, itemapp.ErrUnavailable
	}
	result := itemapp.Snapshot{ID: id, TenantID: tenantID, FolderID: folderID,
		OwnerID: ownerID, CurrentVersionID: currentID, State: state,
		RowVersion: wire.RowVersion, Description: wire.Description,
		Tags: append([]string(nil), wire.Tags...), CreatedAt: wire.CreatedAt.UTC(),
		UpdatedAt: wire.UpdatedAt.UTC(), Current: itemapp.Version{ID: versionID,
			Number: wire.Current.Number, OriginalName: wire.Current.OriginalName,
			MIME: wire.Current.MIME, FileKind: strings.ToLower(wire.Current.FileKind),
			PlaintextBytes: wire.Current.PlaintextBytes,
			CreatedAt:      wire.Current.CreatedAt.UTC()}}
	copy(result.Current.PlaintextSHA256[:], digest)
	copyOptionalTimes(&result, wire)
	return result, nil
}

func copyOptionalTimes(result *itemapp.Snapshot, wire itemDocument) {
	for source, target := range map[*time.Time]*time.Time{wire.TrashedAt: &result.TrashedAt,
		wire.RetentionEndsAt:  &result.RetentionEndsAt,
		wire.PurgeRequestedAt: &result.PurgeRequestedAt,
		wire.PurgeStartsAt:    &result.PurgeStartsAt} {
		if source != nil {
			*target = source.UTC()
		}
	}
}

func decodeState(value string) (domain.ItemState, bool) {
	states := map[string]domain.ItemState{"ACTIVE": domain.ItemActive,
		"TRASHED": domain.ItemTrashed, "PURGE_COOLING": domain.ItemPurgeCooling,
		"PURGING": domain.ItemPurging, "PURGED": domain.ItemPurged}
	state, ok := states[value]
	return state, ok
}
