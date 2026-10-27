package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrFolderInvalid    = errors.New("documents folder input is invalid")
	ErrFolderTransition = errors.New("documents folder transition is invalid")
)

type FolderState string

const (
	FolderActive  FolderState = "active"
	FolderTrashed FolderState = "trashed"
)

type Folder struct {
	ID, TenantID, ParentID, OwnerID ID
	DisplayName                     string
	State                           FolderState
	RowVersion                      uint64
	CreatedAt, UpdatedAt, TrashedAt time.Time
}

func NewFolder(id, tenantID, parentID, ownerID ID, name string, now time.Time) (Folder, error) {
	name = strings.TrimSpace(name)
	if !id.UUIDv7() || !tenantID.Valid() || parentID.Valid() && !parentID.UUIDv7() ||
		!ownerID.Valid() || parentID == id || !validFolderName(name) || !validInstant(now) {
		return Folder{}, ErrFolderInvalid
	}
	now = now.UTC()
	return Folder{ID: id, TenantID: tenantID, ParentID: parentID, OwnerID: ownerID,
		DisplayName: name, State: FolderActive, RowVersion: 1,
		CreatedAt: now, UpdatedAt: now}, nil
}

func (folder Folder) Rename(expected uint64, name string, now time.Time) (Folder, error) {
	name = strings.TrimSpace(name)
	if !folder.mutable(expected, now) || !validFolderName(name) {
		return Folder{}, ErrFolderTransition
	}
	folder.DisplayName, folder.UpdatedAt = name, now.UTC()
	folder.RowVersion++
	return folder, nil
}

func (folder Folder) Move(expected uint64, parent ID, now time.Time) (Folder, error) {
	if !folder.mutable(expected, now) || parent.Valid() && !parent.UUIDv7() || parent == folder.ID {
		return Folder{}, ErrFolderTransition
	}
	folder.ParentID, folder.UpdatedAt = parent, now.UTC()
	folder.RowVersion++
	return folder, nil
}

func (folder Folder) Trash(expected uint64, now time.Time) (Folder, error) {
	if !folder.mutable(expected, now) {
		return Folder{}, ErrFolderTransition
	}
	folder.State, folder.TrashedAt, folder.UpdatedAt = FolderTrashed, now.UTC(), now.UTC()
	folder.RowVersion++
	return folder, nil
}

func (folder Folder) Restore(expected uint64, now time.Time) (Folder, error) {
	if folder.State != FolderTrashed || folder.RowVersion != expected || !validInstant(now) {
		return Folder{}, ErrFolderTransition
	}
	folder.State, folder.TrashedAt, folder.UpdatedAt = FolderActive, time.Time{}, now.UTC()
	folder.RowVersion++
	return folder, nil
}

func (folder Folder) mutable(expected uint64, now time.Time) bool {
	return folder.ID.UUIDv7() && folder.State == FolderActive && folder.RowVersion == expected &&
		expected > 0 && validInstant(now)
}

func validFolderName(value string) bool {
	if value == "" || len([]byte(value)) > 1024 || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}
