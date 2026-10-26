package folder

import (
	"encoding/json"
	"time"

	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type folderDocument struct {
	ID          string     `json:"id"`
	ParentID    string     `json:"parent_id"`
	OwnerID     string     `json:"owner_id"`
	DisplayName string     `json:"display_name"`
	State       string     `json:"state"`
	RowVersion  uint64     `json:"row_version"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	TrashedAt   *time.Time `json:"trashed_at"`
}

func decodeFolder(tenantID domain.ID, document []byte) (domain.Folder, error) {
	var wire folderDocument
	if json.Unmarshal(document, &wire) != nil {
		return domain.Folder{}, folderapp.ErrUnavailable
	}
	id, idErr := domain.ParseID(wire.ID)
	ownerID, ownerErr := domain.ParseID(wire.OwnerID)
	var parentID domain.ID
	var parentErr error
	if wire.ParentID != "" {
		parentID, parentErr = domain.ParseID(wire.ParentID)
	}
	state, stateOK := decodeState(wire.State)
	if idErr != nil || ownerErr != nil || parentErr != nil || !stateOK ||
		wire.RowVersion == 0 || wire.CreatedAt.IsZero() || wire.UpdatedAt.IsZero() {
		return domain.Folder{}, folderapp.ErrUnavailable
	}
	result := domain.Folder{ID: id, TenantID: tenantID, ParentID: parentID, OwnerID: ownerID,
		DisplayName: wire.DisplayName, State: state, RowVersion: wire.RowVersion,
		CreatedAt: wire.CreatedAt.UTC(), UpdatedAt: wire.UpdatedAt.UTC()}
	if wire.TrashedAt != nil {
		result.TrashedAt = wire.TrashedAt.UTC()
	}
	return result, nil
}

func decodeState(value string) (domain.FolderState, bool) {
	switch value {
	case "ACTIVE":
		return domain.FolderActive, true
	case "TRASHED":
		return domain.FolderTrashed, true
	default:
		return "", false
	}
}

func scanMutation(value row, tenantID domain.ID, notFound rowNotFound) (domain.Folder, bool, error) {
	var document []byte
	var replay bool
	if err := value.Scan(&document, &replay); err != nil {
		if notFound(err) {
			return domain.Folder{}, false, folderapp.ErrConflict
		}
		return domain.Folder{}, false, folderapp.ErrUnavailable
	}
	result, err := decodeFolder(tenantID, document)
	return result, replay, err
}

func scanFolder(value row, tenantID domain.ID, notFound rowNotFound) (domain.Folder, error) {
	var document []byte
	if err := value.Scan(&document); err != nil {
		if notFound(err) {
			return domain.Folder{}, folderapp.ErrNotFound
		}
		return domain.Folder{}, folderapp.ErrUnavailable
	}
	return decodeFolder(tenantID, document)
}
