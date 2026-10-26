package documents

import (
	"encoding/hex"
	"time"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

type httpFolderView struct {
	ID         string  `json:"id"`
	ParentID   *string `json:"parent_id"`
	Name       string  `json:"name"`
	State      string  `json:"state"`
	RowVersion uint64  `json:"row_version"`
}

type httpVersionView struct {
	ID           string    `json:"id"`
	Number       uint64    `json:"number"`
	OriginalName string    `json:"original_name"`
	MIME         string    `json:"mime"`
	Bytes        int64     `json:"bytes"`
	SHA256       string    `json:"sha256"`
	CreatedAt    time.Time `json:"created_at"`
}

type httpItemView struct {
	ID               string   `json:"id"`
	FolderID         string   `json:"folder_id"`
	State            string   `json:"state"`
	RowVersion       uint64   `json:"row_version"`
	CurrentVersionID string   `json:"current_version_id"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
}

type httpUploadView struct {
	ID                 string    `json:"id"`
	DocumentID         string    `json:"document_id"`
	CandidateVersionID string    `json:"candidate_version_id"`
	State              string    `json:"state"`
	AttemptGeneration  uint64    `json:"attempt_generation"`
	ExpectedBytes      int64     `json:"expected_bytes"`
	StableErrorCode    string    `json:"stable_error_code,omitempty"`
	ExpiresAt          time.Time `json:"expires_at"`
}

type httpArchiveView struct {
	ID             string     `json:"id"`
	JobID          string     `json:"job_id"`
	State          string     `json:"state"`
	EntryCount     int        `json:"entry_count"`
	PlaintextBytes int64      `json:"plaintext_bytes"`
	ReadyAt        *time.Time `json:"ready_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func httpFolder(folder folderapp.Folder) httpFolderView {
	result := httpFolderView{ID: folder.ID.String(), Name: folder.DisplayName,
		State: string(folder.State), RowVersion: folder.RowVersion}
	if folder.ParentID.Valid() {
		value := folder.ParentID.String()
		result.ParentID = &value
	}
	return result
}

func httpVersion(version itemapp.Version) httpVersionView {
	return httpVersionView{ID: version.ID.String(), Number: version.Number,
		OriginalName: version.OriginalName, MIME: version.MIME,
		Bytes: version.PlaintextBytes, SHA256: hex.EncodeToString(version.PlaintextSHA256[:]),
		CreatedAt: version.CreatedAt.UTC()}
}

func httpItem(item itemapp.Snapshot) httpItemView {
	return httpItemView{ID: item.ID.String(), FolderID: item.FolderID.String(),
		State: string(item.State), RowVersion: item.RowVersion,
		CurrentVersionID: item.CurrentVersionID.String(), Description: item.Description,
		Tags: append([]string(nil), item.Tags...)}
}

func httpUpload(upload uploadapp.Snapshot) httpUploadView {
	return httpUploadView{ID: upload.ID.String(), DocumentID: upload.DocumentID.String(),
		CandidateVersionID: upload.CandidateVersionID.String(), State: string(upload.State),
		AttemptGeneration: upload.AttemptGeneration, ExpectedBytes: upload.ExpectedBytes,
		StableErrorCode: upload.StableErrorCode, ExpiresAt: upload.ExpiresAt.UTC()}
}

func httpArchive(archive archiveapp.Snapshot) httpArchiveView {
	result := httpArchiveView{ID: archive.ID.String(), JobID: archive.JobID.String(),
		State: string(archive.State), EntryCount: archive.EntryCount,
		PlaintextBytes: archive.PlaintextBytes, CreatedAt: archive.CreatedAt.UTC()}
	if !archive.ReadyAt.IsZero() {
		value := archive.ReadyAt.UTC()
		result.ReadyAt = &value
	}
	if !archive.ExpiresAt.IsZero() {
		value := archive.ExpiresAt.UTC()
		result.ExpiresAt = &value
	}
	return result
}
