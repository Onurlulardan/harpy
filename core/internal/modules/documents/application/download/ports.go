package download

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	LoadCurrent(context.Context, domain.ID) (Descriptor, error)
	LoadVersion(context.Context, domain.ID, domain.ID) (Descriptor, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope) (Store, foundation.Identity, error)
}

type Keyring interface {
	Unwrap(context.Context, Descriptor) ([32]byte, error)
}

type Cipher interface {
	Decrypt(context.Context, io.Reader, io.Writer, [32]byte,
		CipherMetadata) (CipherReceipt, error)
}

type ObjectStore interface {
	OpenExact(context.Context, ObjectExpected) (io.ReadCloser, error)
}

type ScopeRunner interface {
	InSnapshot(context.Context, string,
		func(context.Context, platformtx.TenantScope) error) error
}
