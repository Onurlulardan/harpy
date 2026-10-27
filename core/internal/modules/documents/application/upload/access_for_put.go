package upload

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) bindPut(ctx context.Context, scope platformtx.TenantScope,
	uploadID domain.ID, operation, permission string,
) (Store, Snapshot, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		scope.OperationID() != operation || !uploadID.UUIDv7() {
		return nil, Snapshot{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope, Command)
	if err != nil || store == nil || !identity.Valid() {
		return nil, Snapshot{}, ErrUnavailable
	}
	snapshot, err := store.Get(ctx, uploadID)
	if err != nil {
		return nil, Snapshot{}, mapStoreError(err)
	}
	if permission == "" {
		permission = PermissionCreate
		if snapshot.ExpectedCurrentVersionID.Valid() {
			permission = PermissionVersion
		}
	}
	if err = application.authorizer.Authorize(ctx, scope, permission); err != nil {
		return nil, Snapshot{}, ErrForbidden
	}
	return store, snapshot, nil
}

func putPermission(plan AttemptPlan, override string) string {
	if override != "" {
		return override
	}
	if plan.ExpectedCurrentVersionID.Valid() {
		return PermissionVersion
	}
	return PermissionCreate
}

func replayResult(snapshot Snapshot) (ReadyResult, bool, error) {
	switch snapshot.State {
	case domain.UploadReady:
		if snapshot.ReadyStatus != 201 || len(snapshot.ReadyResponse) == 0 {
			return ReadyResult{}, true, ErrUnavailable
		}
		return ReadyResult{Upload: snapshot, DocumentID: snapshot.DocumentID,
			VersionID: snapshot.CandidateVersionID, Status: snapshot.ReadyStatus,
			Response: append([]byte(nil), snapshot.ReadyResponse...)}, true, nil
	case domain.UploadCreated, domain.UploadRetryable:
		return ReadyResult{}, false, nil
	case domain.UploadReceiving, domain.UploadScanning:
		return ReadyResult{}, true, ErrInProgress
	case domain.UploadRejected:
		return ReadyResult{}, true, ErrContentRejected
	default:
		return ReadyResult{}, true, ErrConflict
	}
}
