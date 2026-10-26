package archivedownload

import (
	"context"
	"fmt"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const Operation = "documents.archive.download"

type Application struct {
	binder     StoreBinder
	authorizer Authorizer
	keyring    Keyring
	cipher     Cipher
	objects    ObjectStore
	clock      foundation.Clock
}

func NewApplication(binder StoreBinder, authorizer Authorizer, keyring Keyring,
	cipher Cipher, objects ObjectStore, clock foundation.Clock,
) (*Application, error) {
	for _, dependency := range []any{binder, authorizer, keyring, cipher, objects, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents archive download dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer, keyring: keyring,
		cipher: cipher, objects: objects, clock: clock}, nil
}

func (application *Application) descriptor(ctx context.Context, runner ScopeRunner,
	id domain.ID,
) (archiveapp.Plan, foundation.Identity, uint64, error) {
	var plan archiveapp.Plan
	var identity foundation.Identity
	var revision uint64
	err := runner.InSnapshot(ctx, Operation, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		var authorizeErr error
		revision, authorizeErr = application.authorizer.Authorize(scoped, scope,
			archiveapp.PermissionRead)
		if authorizeErr != nil || revision == 0 {
			return ErrForbidden
		}
		store, boundIdentity, bindErr := application.binder.Bind(scoped, scope,
			archiveapp.SnapshotAccess)
		if bindErr != nil || store == nil || !boundIdentity.Valid() {
			return ErrUnavailable
		}
		identity = boundIdentity
		plan, bindErr = store.LoadPlan(scoped, id)
		return bindErr
	})
	return plan, identity, revision, mapError(err)
}

func archiveDescriptor(plan archiveapp.Plan) downloadapp.Descriptor {
	return downloadapp.Descriptor{TenantID: plan.TenantID, ItemID: plan.ID,
		VersionID: plan.ID, OriginalName: "archive.zip", MIME: "application/zip",
		Object: domain.ObjectReceipt{Key: plan.ObjectKey, Version: plan.ObjectVersion,
			KeyID: plan.KeyID, Generation: plan.ObjectGeneration,
			PlaintextBytes:   plan.ArtifactPlaintextBytes,
			CiphertextBytes:  plan.CiphertextBytes,
			PlaintextSHA256:  plan.ArtifactSHA256,
			CiphertextSHA256: plan.CiphertextSHA256,
			WrappedDEK:       append([]byte(nil), plan.WrappedDEK...)}}
}
