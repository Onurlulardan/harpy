package download

import (
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Request struct {
	ItemID, VersionID domain.ID
}

type Descriptor struct {
	TenantID, ItemID, VersionID domain.ID
	OriginalName, MIME          string
	Object                      domain.ObjectReceipt
	CreatedAt                   time.Time
}

type Result struct {
	ItemID, VersionID domain.ID
	Filename, MIME    string
	Bytes             int64
	SHA256            [32]byte
	Content           io.ReadCloser
}

type CipherMetadata struct {
	TenantID, ItemID, VersionID domain.ID
	Generation                  uint64
	PlaintextBytes              int64
}

type CipherReceipt struct {
	PlaintextSHA256, CiphertextSHA256 [32]byte
}

type ObjectExpected struct {
	Key, Version     string
	Generation       uint64
	CiphertextBytes  int64
	CiphertextSHA256 [32]byte
}
