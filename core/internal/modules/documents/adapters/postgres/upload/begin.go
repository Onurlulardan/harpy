package upload

import (
	"context"

	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	"github.com/google/uuid"
)

func (storage *store) Begin(ctx context.Context, uploadID, leaseOwnerID,
	effectReceiptID domain.ID,
) (uploadapp.AttemptPlan, error) {
	if !storage.usable(ctx) || !storage.command || !uploadID.UUIDv7() ||
		!leaseOwnerID.UUIDv7() || !effectReceiptID.UUIDv7() {
		return uploadapp.AttemptPlan{}, uploadapp.ErrInvalid
	}
	var id, actor, document, candidate, reservation uuid.UUID
	var expected *uuid.UUID
	var generation int64
	var name, mime string
	var digest []byte
	var expectedBytes int64
	err := storage.queryRow(ctx, beginSQL, uploadID.String(), storage.actorID.String(),
		leaseOwnerID.String(), effectReceiptID.String()).Scan(&id, &actor, &document,
		&candidate, &expected, &reservation, &generation, &name, &mime, &digest,
		&expectedBytes)
	if err != nil {
		if storage.notFound(err) {
			return uploadapp.AttemptPlan{}, uploadapp.ErrConflict
		}
		return uploadapp.AttemptPlan{}, uploadapp.ErrUnavailable
	}
	kind, err := domain.DeclaredFileKind(name, mime, expectedBytes)
	key, keyErr := domain.ContentObjectKey(storage.tenantID, domain.ID(document),
		domain.ID(candidate))
	if err != nil || keyErr != nil || len(digest) != 32 || generation < 1 ||
		id == uuid.Nil || actor != uuid.UUID(storage.actorID) {
		return uploadapp.AttemptPlan{}, uploadapp.ErrUnavailable
	}
	plan := uploadapp.AttemptPlan{UploadID: domain.ID(id), TenantID: storage.tenantID,
		ActorID: domain.ID(actor), DocumentID: domain.ID(document),
		CandidateVersionID: domain.ID(candidate), BillingReservationID: domain.ID(reservation),
		EffectReceiptID: effectReceiptID, Generation: uint64(generation), ObjectKey: key,
		OriginalName: name, DeclaredMIME: mime, ExpectedBytes: expectedBytes,
		DeclaredKind: kind}
	if expected != nil {
		plan.ExpectedCurrentVersionID = domain.ID(*expected)
	}
	copy(plan.ExpectedContentSHA256[:], digest)
	return plan, nil
}
