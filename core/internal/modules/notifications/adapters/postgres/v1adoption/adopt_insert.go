package v1adoption

import (
	"context"
	"errors"
	"fmt"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
)

func (store *Store) insertMapped(ctx context.Context, value app.MappedRecord) error {
	if err := store.execute(ctx, upsertCounterSQL,
		value.RecipientID, value.InboxRevision); err != nil {
		return errors.Join(app.ErrUnavailable, fmt.Errorf("advance notification inbox counter: %w", err))
	}
	if err := store.execute(ctx, insertMessageSQL, value.MessageID, value.RecipientID,
		value.NotificationType, value.Category, value.Subject, value.Body,
		value.CanonicalDigest[:], value.SubjectDigest[:], value.BodyDigest[:],
		value.InboxRevision, value.CreatedAt, value.ReadAt); err != nil {
		return errors.Join(app.ErrConflict, fmt.Errorf("insert adopted notification message: %w", err))
	}
	if err := store.execute(ctx, insertEffectSQL, value.EffectKeyHash[:],
		value.RecipientID, value.CanonicalDigest[:], value.MessageID, value.CreatedAt); err != nil {
		return errors.Join(app.ErrConflict, fmt.Errorf("insert adopted notification effect: %w", err))
	}
	if value.Channel == "" {
		return nil
	}
	if err := store.execute(ctx, insertDeliverySQL,
		value.DeliveryID, value.MessageID, value.CreatedAt); err != nil {
		return errors.Join(app.ErrConflict, fmt.Errorf("insert adopted notification delivery: %w", err))
	}
	if err := store.execute(ctx, insertAttemptSQL, value.AttemptID,
		value.DeliveryID, value.ProviderReferenceDigest[:], value.CreatedAt); err != nil {
		return errors.Join(app.ErrConflict, fmt.Errorf("insert adopted notification attempt: %w", err))
	}
	return nil
}
