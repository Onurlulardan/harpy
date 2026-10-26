package archiveartifact

import (
	"context"
	"fmt"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Writer struct {
	contents ContentSource
	keyring  Keyring
	cipher   Cipher
	objects  ObjectStore
	clock    Clock
}

func New(contents ContentSource, keyring Keyring, cipher Cipher,
	objects ObjectStore, clock Clock,
) (*Writer, error) {
	for _, dependency := range []any{contents, keyring, cipher, objects, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents archive artifact dependencies are invalid")
		}
	}
	return &Writer{contents: contents, keyring: keyring, cipher: cipher,
		objects: objects, clock: clock}, nil
}

func (writer *Writer) UpperBound(plan archiveapp.Plan) (int64, error) {
	if writer == nil || !validPlan(plan) {
		return 0, archiveapp.ErrInvalid
	}
	return zipSize(plan.Entries)
}

func (writer *Writer) Write(ctx context.Context, plan archiveapp.Plan,
	refresh func(int) error,
) (archiveapp.ArtifactReceipt, error) {
	if writer == nil || ctx == nil || ctx.Err() != nil || refresh == nil ||
		!foundation.ExternalEffectsAllowed(ctx) || !validPlan(plan) || plan.ObjectVersion != "" {
		return archiveapp.ArtifactReceipt{}, archiveapp.ErrInvalid
	}
	plaintextBytes, err := zipSize(plan.Entries)
	if err != nil {
		return archiveapp.ArtifactReceipt{}, err
	}
	key, err := writer.keyring.CreateDataKey(ctx, plan.TenantID, plan.ID, plan.ID)
	if err != nil || key.KeyID == "" || len(key.Wrapped) < 115 {
		clear(key.Plaintext[:])
		return archiveapp.ArtifactReceipt{}, archiveapp.ErrUnavailable
	}
	defer clear(key.Plaintext[:])
	objectKey, err := domain.ArchiveObjectKey(plan.TenantID, plan.ID)
	if err != nil {
		return archiveapp.ArtifactReceipt{}, archiveapp.ErrInvalid
	}
	cipherReceipt, object, err := writer.stream(ctx, plan, refresh, plaintextBytes,
		objectKey, key.Plaintext)
	if err != nil {
		return archiveapp.ArtifactReceipt{}, err
	}
	return archiveapp.ArtifactReceipt{ObjectKey: objectKey, ObjectVersion: object.Version,
		ObjectGeneration: object.Generation, PlaintextBytes: plaintextBytes,
		CiphertextBytes:  cipherReceipt.CiphertextBytes,
		PlaintextSHA256:  cipherReceipt.PlaintextSHA256,
		CiphertextSHA256: cipherReceipt.CiphertextSHA256,
		KeyID:            key.KeyID, WrappedDEK: append([]byte(nil), key.Wrapped...),
		WrittenAt: writer.clock.Now().UTC()}, nil
}

func validPlan(plan archiveapp.Plan) bool {
	return plan.ID.UUIDv7() && plan.TenantID.Valid() && plan.ActorID.Valid() &&
		plan.JobID.UUIDv7() && plan.BillingReservationID.UUIDv7() &&
		plan.AuthorizationRevision > 0 && len(plan.Entries) > 0 &&
		len(plan.Entries) <= domain.MaximumArchiveFiles
}

var _ archiveapp.ArtifactWriter = (*Writer)(nil)
