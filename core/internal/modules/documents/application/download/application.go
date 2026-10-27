package download

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const (
	OperationCurrent = "documents.item.download"
	OperationVersion = "documents.item.version.download"
	Permission       = "documents.item.download"
)

type Application struct {
	binder     StoreBinder
	authorizer foundation.Authorizer
	keyring    Keyring
	cipher     Cipher
	objects    ObjectStore
}

func NewApplication(binder StoreBinder, authorizer foundation.Authorizer,
	keyring Keyring, cipher Cipher, objects ObjectStore,
) (*Application, error) {
	for _, dependency := range []any{binder, authorizer, keyring, cipher, objects} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents download dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer,
		keyring: keyring, cipher: cipher, objects: objects}, nil
}

func (application *Application) descriptor(ctx context.Context, runner ScopeRunner,
	request Request, operation, permission string,
) (Descriptor, error) {
	var result Descriptor
	err := runner.InSnapshot(ctx, operation, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		if scope.OperationID() != operation ||
			application.authorizer.Authorize(scoped, scope, permission) != nil {
			return ErrForbidden
		}
		store, identity, bindErr := application.binder.Bind(scoped, scope)
		if bindErr != nil || store == nil || !identity.Valid() {
			return ErrUnavailable
		}
		if request.VersionID.Valid() {
			result, bindErr = store.LoadVersion(scoped, request.ItemID, request.VersionID)
		} else {
			result, bindErr = store.LoadCurrent(scoped, request.ItemID)
		}
		return bindErr
	})
	return result, mapError(err)
}
