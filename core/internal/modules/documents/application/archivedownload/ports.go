package archivedownload

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type StoreBinder = archiveapp.StoreBinder
type Authorizer interface {
	archiveapp.CommandAuthorizer
	archiveapp.RefreshAuthorizer
}
type Keyring = downloadapp.Keyring
type Cipher = downloadapp.Cipher
type ObjectStore = downloadapp.ObjectStore

type ScopeRunner interface {
	InSnapshot(context.Context, string,
		func(context.Context, platformtx.TenantScope) error) error
}
