package archivedownload

import (
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Request struct{ ArchiveID domain.ID }

type Result struct {
	ArchiveID domain.ID
	Filename  string
	Bytes     int64
	SHA256    [32]byte
	Content   io.ReadCloser
}
