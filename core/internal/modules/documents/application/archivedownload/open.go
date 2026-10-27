package archivedownload

import (
	"context"
	"fmt"

	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Open(ctx context.Context, runner ScopeRunner,
	request Request,
) (Result, error) {
	if application == nil || foundation.NilPort(runner) || !request.ArchiveID.UUIDv7() ||
		!platformtx.NetworkAllowed(ctx) {
		return Result{}, ErrInvalid
	}
	plan, identity, revision, err := application.descriptor(ctx, runner, request.ArchiveID)
	if err != nil {
		return Result{}, err
	}
	if plan.ID != request.ArchiveID || plan.State != domain.ArchiveReady ||
		!application.clock.Now().Before(plan.ExpiresAt) || len(plan.Entries) == 0 {
		if plan.State == domain.ArchiveExpired || plan.State == domain.ArchivePurged ||
			!plan.ExpiresAt.IsZero() && !application.clock.Now().Before(plan.ExpiresAt) {
			return Result{}, ErrExpired
		}
		return Result{}, ErrNotFound
	}
	itemIDs := make([]domain.ID, len(plan.Entries))
	for index, entry := range plan.Entries {
		itemIDs[index] = entry.DocumentID
	}
	if err = application.authorizer.Refresh(ctx, plan.TenantID, identity.ActorID,
		revision, itemIDs); err != nil {
		return Result{}, ErrForbidden
	}
	descriptor := archiveDescriptor(plan)
	key, err := application.keyring.Unwrap(ctx, descriptor)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	expected := downloadapp.ObjectExpected{Key: plan.ObjectKey, Version: plan.ObjectVersion,
		Generation: plan.ObjectGeneration, CiphertextBytes: plan.CiphertextBytes,
		CiphertextSHA256: plan.CiphertextSHA256}
	object, err := application.objects.OpenExact(ctx, expected)
	if err != nil {
		clear(key[:])
		return Result{}, mapError(err)
	}
	content := application.decrypt(ctx, object, key, descriptor)
	return Result{ArchiveID: plan.ID,
		Filename: fmt.Sprintf("archive-%s.zip", plan.ID.String()),
		Bytes:    plan.ArtifactPlaintextBytes, SHA256: plan.ArtifactSHA256,
		Content: content}, nil
}
