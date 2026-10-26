package knowledgefile

import (
	"context"
	"fmt"
	"time"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Stage(ctx context.Context,
	request knowledgeapi.StageRequest,
) (knowledgeapi.Staged, error) {
	if application == nil || !platformtx.NetworkAllowed(ctx) {
		return knowledgeapi.Staged{}, knowledgeapi.ErrInvalid
	}
	access, err := stageAccess(request, application.clock.Now().UTC())
	if err != nil {
		return knowledgeapi.Staged{}, knowledgeapi.ErrInvalid
	}
	folderID, err := application.ensureFolder(ctx, access)
	if err != nil {
		return knowledgeapi.Staged{}, mapError(err)
	}
	var upload uploadapp.Snapshot
	err = application.scopes.InCommand(ctx, access, uploadapp.OperationCreate,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var inner error
			upload, inner = application.uploads.Create(scoped, scope, uploadapp.CreateRequest{
				FolderID: folderID, OriginalName: request.Filename, DeclaredMIME: request.MIME,
				IdempotencyKey: "ai-knowledge-stage-" + request.OperationID.String(),
				Description:    "AI Knowledge Base upload", ExpectedContentSHA256: request.SHA256,
				ExpectedBytes: request.Bytes})
			if inner != nil {
				return inner
			}
			store, identity, bindErr := application.leases.Bind(scoped, scope)
			if bindErr != nil || identity.TenantID != access.TenantID ||
				identity.ActorID != access.ActorID {
				return errUnavailable
			}
			_, inner = store.Create(scoped, StageSeed{Lease: Lease{
				LeaseID: upload.ID, UploadID: upload.ID, TenantID: access.TenantID,
				ActorID: access.ActorID, ItemID: upload.DocumentID,
				VersionID: upload.CandidateVersionID, ExpectedBytes: request.Bytes,
				ExpectedSHA256: request.SHA256, Filename: request.Filename,
				MIME: request.MIME, State: "PENDING", ExpiresAt: upload.ExpiresAt},
				CreatedAt: application.clock.Now().UTC()})
			return inner
		})
	if err != nil {
		return knowledgeapi.Staged{}, mapError(err)
	}
	ready, err := application.uploads.Put(ctx, boundRunner{application, access},
		uploadapp.PutRequest{UploadID: upload.ID,
			LeaseOwnerID: domain.ID(request.OperationID), Content: request.Content})
	if err != nil {
		return knowledgeapi.Staged{}, mapError(err)
	}
	if ready.DocumentID != upload.DocumentID || ready.VersionID != upload.CandidateVersionID {
		return knowledgeapi.Staged{}, knowledgeapi.ErrUnavailable
	}
	return application.completeStage(ctx, access, upload.ID)
}

func (application *Application) ensureFolder(ctx context.Context,
	access Access,
) (domain.ID, error) {
	var result domain.ID
	err := application.scopes.InCommand(ctx, access, folderapp.OperationCreate,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			folder, inner := application.folders.Create(scoped, scope, folderapp.CreateRequest{
				DisplayName: knowledgeFolderName, IdempotencyKey: knowledgeFolderKey})
			result = folder.ID
			return inner
		})
	return result, err
}

func (application *Application) completeStage(ctx context.Context, access Access,
	leaseID domain.ID,
) (knowledgeapi.Staged, error) {
	var lease Lease
	err := application.scopes.InCommand(ctx, access, uploadapp.OperationPut,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			store, identity, bindErr := application.leases.Bind(scoped, scope)
			if bindErr != nil || identity.TenantID != access.TenantID ||
				identity.ActorID != access.ActorID {
				return errUnavailable
			}
			lease, bindErr = store.CompleteStage(scoped, leaseID,
				application.clock.Now().UTC())
			return bindErr
		})
	if err != nil {
		return knowledgeapi.Staged{}, mapError(err)
	}
	return stagedWire(lease), nil
}

func stageAccess(request knowledgeapi.StageRequest, now time.Time) (Access, error) {
	if !request.OperationID.UUIDv7() || !request.TenantID.Valid() ||
		!request.ActorID.Valid() || request.AuthorizationVersion < 1 ||
		request.Filename == "" || request.MIME == "" || request.Bytes < 1 ||
		request.SHA256 == [32]byte{} || request.Content == nil ||
		!foundation.ValidInstant(now) ||
		!request.ExpiresAt.After(now) || request.ExpiresAt.After(now.Add(2*time.Hour)) {
		return Access{}, fmt.Errorf("invalid stage request")
	}
	return Access{TenantID: domain.ID(request.TenantID), ActorID: domain.ID(request.ActorID),
		AuthorizationVersion: request.AuthorizationVersion}, nil
}
