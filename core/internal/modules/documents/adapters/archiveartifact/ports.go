package archiveartifact

import (
	"context"
	"io"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type ContentSource interface {
	Open(context.Context, archiveapp.Entry) (io.ReadCloser, error)
}

type Keyring interface {
	CreateDataKey(context.Context, domain.ID, domain.ID,
		domain.ID) (uploadapp.DataKey, error)
}

type Cipher interface {
	Encrypt(context.Context, io.Reader, io.Writer, [32]byte,
		uploadapp.CipherMetadata) (uploadapp.CipherReceipt, error)
}

type ObjectStore interface {
	Put(context.Context, uploadapp.ObjectPutRequest) (uploadapp.ObjectPutResult, error)
	Reconcile(context.Context, uploadapp.ObjectExpected) (uploadapp.ObjectPutResult, error)
	Verify(context.Context, uploadapp.ObjectExpected) error
}

type Clock = foundation.Clock
