package copy

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Copy(ctx context.Context, runner ScopeRunner,
	request Request,
) (Result, error) {
	request.OriginalName = strings.TrimSpace(request.OriginalName)
	if application == nil || foundation.NilPort(runner) ||
		!request.SourceItemID.UUIDv7() ||
		(request.SourceVersionID.Valid() && !request.SourceVersionID.UUIDv7()) ||
		!request.TargetFolderID.UUIDv7() ||
		!foundation.ValidIdempotency(request.IdempotencyKey) ||
		!platformtx.NetworkAllowed(ctx) {
		return Result{}, ErrInvalid
	}
	if err := application.checkAccess(ctx, runner); err != nil {
		return Result{}, err
	}
	source, err := application.dependencies.Source.Open(ctx, runner,
		request.SourceItemID, request.SourceVersionID)
	if err != nil {
		return Result{}, mapError(err)
	}
	if !validSource(source, request) {
		if source.Content != nil {
			_ = source.Content.Close()
		}
		return Result{}, ErrUnavailable
	}
	defer source.Content.Close()
	leaseID, err := application.dependencies.LeaseIDs.NewUUIDv7()
	if err != nil || !leaseID.UUIDv7() {
		return Result{}, ErrUnavailable
	}
	name := request.OriginalName
	if name == "" {
		name = source.OriginalName
	}
	result, err := application.dependencies.Destination.CreateAndPut(ctx, runner,
		DestinationRequest{FolderID: request.TargetFolderID, LeaseOwnerID: leaseID,
			OriginalName: name, MIME: source.MIME, Description: request.Description,
			Tags: request.Tags, Bytes: source.Bytes, SHA256: source.SHA256,
			IdempotencyKey: request.IdempotencyKey, Content: source.Content})
	if err != nil {
		return Result{}, mapError(err)
	}
	if !validResult(result, source) {
		return Result{}, ErrUnavailable
	}
	result.SourceItemID, result.SourceVersionID = source.ItemID, source.VersionID
	return result, nil
}

func (application *Application) checkAccess(ctx context.Context, runner ScopeRunner) error {
	return mapError(runner.InCommand(ctx, OperationCopy, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		if scope.OperationID() != OperationCopy {
			return ErrForbidden
		}
		return application.dependencies.Authorizer.Authorize(scoped, scope, PermissionCopy)
	}))
}

func validSource(source Source, request Request) bool {
	return source.ItemID == request.SourceItemID && source.VersionID.UUIDv7() &&
		(!request.SourceVersionID.Valid() || source.VersionID == request.SourceVersionID) &&
		source.OriginalName != "" && source.MIME != "" && source.Bytes > 0 &&
		source.SHA256 != [32]byte{} && source.Content != nil
}

func validResult(result Result, source Source) bool {
	return result.ItemID.UUIDv7() && result.VersionID.UUIDv7() && result.UploadID.UUIDv7() &&
		result.ItemID != source.ItemID && result.VersionID != source.VersionID &&
		result.Bytes == source.Bytes && result.SHA256 == source.SHA256
}
