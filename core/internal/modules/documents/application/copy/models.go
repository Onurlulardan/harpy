package copy

import (
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Request struct {
	SourceItemID, SourceVersionID domain.ID
	TargetFolderID                domain.ID
	OriginalName                  string
	Description                   string
	Tags                          []string
	IdempotencyKey                string
}

type Source struct {
	ItemID, VersionID  domain.ID
	OriginalName, MIME string
	Bytes              int64
	SHA256             [32]byte
	Content            io.ReadCloser
}

type DestinationRequest struct {
	FolderID, LeaseOwnerID domain.ID
	OriginalName, MIME     string
	Description            string
	Tags                   []string
	Bytes                  int64
	SHA256                 [32]byte
	IdempotencyKey         string
	Content                io.Reader
}

type Result struct {
	SourceItemID, SourceVersionID domain.ID
	ItemID, VersionID, UploadID   domain.ID
	Bytes                         int64
	SHA256                        [32]byte
}
