package adoptionartifact

import (
	"context"
	"io"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type SourceStore interface {
	OpenExact(context.Context, string, string) (io.ReadCloser, error)
}

type MetadataVerifier interface {
	VerifyInactive(context.Context, adoptionapp.Checkpoint) ([32]byte, error)
}

type Keyring interface {
	CreateDataKey(context.Context,
		domain.ID, domain.ID, domain.ID) (uploadapp.DataKey, error)
}

type Cipher = uploadapp.Cipher
type ObjectStore = uploadapp.ObjectStore
