package upload

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Create(ctx context.Context, scope platformtx.TenantScope,
	request CreateRequest,
) (Snapshot, error) {
	permission := PermissionCreate
	if request.DocumentID.Valid() {
		permission = PermissionVersion
	}
	return application.create(ctx, scope, request, OperationCreate, permission)
}

func (application *Application) create(ctx context.Context, scope platformtx.TenantScope,
	request CreateRequest, operation, permission string,
) (Snapshot, error) {
	request.OriginalName = strings.TrimSpace(request.OriginalName)
	description, tags, encodedTags, metadataOK := foundation.NormalizeItemMetadata(request.Description,
		request.Tags)
	request.Description, request.Tags = description, tags
	if !metadataOK || !validCreateRequest(request) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, Command, operation, permission)
	if err != nil {
		return Snapshot{}, err
	}
	requestDigest := foundation.Digest([]byte(OperationCreate), identity.ActorID[:],
		identity.TenantID[:], request.FolderID[:], request.DocumentID[:],
		request.ExpectedCurrentVersionID[:], []byte(request.OriginalName),
		[]byte(request.DeclaredMIME), request.ExpectedContentSHA256[:],
		foundation.Int64Bytes(request.ExpectedBytes), []byte(request.Description), encodedTags)
	if existing, found, findErr := store.FindCreate(ctx, request.IdempotencyKey,
		requestDigest); findErr != nil {
		return Snapshot{}, mapStoreError(findErr)
	} else if found {
		return existing, nil
	}
	return application.createBound(ctx, scope, store, identity, request, requestDigest)
}

func (application *Application) createBound(ctx context.Context, scope platformtx.TenantScope,
	store Store, identity foundation.Identity, request CreateRequest, requestDigest [32]byte,
) (Snapshot, error) {
	ids, err := application.nextIDs(5)
	if err != nil {
		return Snapshot{}, err
	}
	uploadID, documentID, candidateID, operationID, auditID := ids[0], ids[1], ids[2], ids[3], ids[4]
	if request.DocumentID.Valid() {
		documentID = request.DocumentID
	}
	now := application.clock.Now()
	upperBound, boundErr := domain.CiphertextUpperBound(request.ExpectedBytes)
	if boundErr != nil || !foundation.ValidInstant(now) {
		return Snapshot{}, ErrInvalid
	}
	expiresAt := now.UTC().Add(UploadKeep)
	reservation, err := application.billing.Reserve(ctx, scope, ReservationRequest{
		SourceOperationID: uploadID, Amount: upperBound, ExpiresAt: expiresAt})
	if err != nil {
		return Snapshot{}, mapBillingError(err)
	}
	if !reservation.Allowed {
		return Snapshot{}, ErrQuota
	}
	if !reservation.ID.UUIDv7() {
		return Snapshot{}, ErrUnavailable
	}
	mutation := CreateMutation{Snapshot: Snapshot{ID: uploadID, TenantID: identity.TenantID,
		ActorID: identity.ActorID, DocumentID: documentID, CandidateVersionID: candidateID,
		ExpectedCurrentVersionID: request.ExpectedCurrentVersionID, FolderID: request.FolderID,
		BillingReservationID: reservation.ID, RequestSHA256: requestDigest,
		ExpectedContentSHA256: request.ExpectedContentSHA256, ExpectedBytes: request.ExpectedBytes,
		CiphertextUpperBound: upperBound, OriginalName: request.OriginalName,
		DeclaredMIME: request.DeclaredMIME, State: domain.UploadCreated,
		Description: request.Description, Tags: append([]string(nil), request.Tags...),
		ExpiresAt: expiresAt, CreatedAt: now.UTC(), UpdatedAt: now.UTC()},
		OperationID: operationID, AuditID: auditID, IdempotencyKey: request.IdempotencyKey}
	result, err := store.Create(ctx, mutation)
	if err != nil {
		return Snapshot{}, mapStoreError(err)
	}
	return result, nil
}

func validCreateRequest(request CreateRequest) bool {
	_, fileErr := domain.DeclaredFileKind(request.OriginalName, request.DeclaredMIME,
		request.ExpectedBytes)
	newDocument := !request.DocumentID.Valid() && !request.ExpectedCurrentVersionID.Valid()
	newVersion := request.DocumentID.UUIDv7() && request.ExpectedCurrentVersionID.UUIDv7()
	return request.FolderID.UUIDv7() && (newDocument || newVersion) &&
		request.ExpectedContentSHA256 != [32]byte{} && fileErr == nil &&
		foundation.ValidIdempotency(request.IdempotencyKey)
}
