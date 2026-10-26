package link

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (registry *FormRecordRegistry) registerFormRecordLink(
	ctx context.Context,
	store Store,
	identity foundation.Identity,
	source Source,
	itemID domain.ID,
	request formrecordlink.Request,
	now time.Time,
) error {
	epoch, err := store.LockActiveEpoch(ctx, itemID)
	if err != nil {
		return err
	}
	ids, err := registry.nextIDs(3)
	if err != nil {
		return err
	}
	value, err := domain.NewLink(ids[0], identity.TenantID, itemID,
		source.ID, source.VersionID, source.Owner, source.Kind,
		formrecordlink.Usage, epoch, now)
	if err != nil {
		return ErrInvalid
	}
	key := formRecordLinkKey("register", request, itemID)
	digest := foundation.Digest([]byte(OperationRegister), itemID[:],
		source.ID[:], source.VersionID[:], []byte(source.Owner),
		[]byte(source.Kind), []byte(formrecordlink.Usage))
	result, _, err := store.Register(ctx, Mutation{Link: value,
		ActorID: identity.ActorID, OperationID: ids[1], AuditID: ids[2],
		Source: source, IdempotencyKey: key, RequestSHA256: digest,
		OccurredAt: now.UTC()})
	if errors.Is(err, ErrConflict) {
		active, probeErr := registry.sourceHasItem(ctx, store, source, itemID)
		if probeErr != nil {
			return probeErr
		}
		if active {
			return nil
		}
	}
	if err != nil {
		return err
	}
	if result.ItemID != itemID || result.ReferenceEpoch != epoch ||
		result.SourceID != source.ID || result.SourceVersionID != source.VersionID ||
		result.SourceOwner != source.Owner || result.SourceKind != source.Kind ||
		result.Usage != formrecordlink.Usage || result.State != domain.LinkActive {
		return ErrUnavailable
	}
	return nil
}

func (registry *FormRecordRegistry) releaseFormRecordLink(
	ctx context.Context,
	store Store,
	identity foundation.Identity,
	source Source,
	link domain.Link,
	request formrecordlink.Request,
	now time.Time,
) error {
	ids, err := registry.nextIDs(2)
	if err != nil {
		return err
	}
	key := formRecordLinkKey("release", request, link.ItemID)
	digest := foundation.Digest([]byte(OperationRelease), link.ID[:],
		source.ID[:], source.VersionID[:])
	result, _, err := store.Release(ctx, Mutation{Link: domain.Link{
		ID: link.ID, TenantID: identity.TenantID,
	}, ActorID: identity.ActorID, OperationID: ids[0], AuditID: ids[1],
		Source: source, IdempotencyKey: key, RequestSHA256: digest,
		OccurredAt: now.UTC()})
	if errors.Is(err, ErrConflict) {
		active, probeErr := registry.sourceHasItem(ctx, store, source, link.ItemID)
		if probeErr != nil {
			return probeErr
		}
		if !active {
			return nil
		}
	}
	if err != nil {
		return err
	}
	if result.ID != link.ID || result.ItemID != link.ItemID ||
		result.State != domain.LinkReleased {
		return ErrUnavailable
	}
	return nil
}

func (registry *FormRecordRegistry) sourceHasItem(
	ctx context.Context,
	store Store,
	source Source,
	itemID domain.ID,
) (bool, error) {
	items, err := listFormRecordLinks(ctx, store, source)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if item.ItemID == itemID && item.State == domain.LinkActive {
			return true, nil
		}
	}
	return false, nil
}

func formRecordLinkKey(
	operation string,
	request formrecordlink.Request,
	itemID domain.ID,
) string {
	digest := foundation.Digest([]byte("documents.form_record_link."+operation),
		[]byte(request.IdempotencyKey), request.DefinitionID[:],
		request.RecordID[:], itemID[:])
	return "forms-record-link-" + hex.EncodeToString(digest[:])
}
