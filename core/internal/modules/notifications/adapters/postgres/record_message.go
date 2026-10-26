package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) RecordMessage(ctx context.Context, value domain.Message) error {
	if !store.usable(ctx, true) || !value.ID.Valid() || !value.RecipientID.Valid() ||
		!value.TemplateVersionID.Valid() {
		return ErrInvalid
	}
	_, err := store.execute(ctx, insertMessageSQL, value.ID.String(),
		value.RecipientID.String(), value.TemplateVersionID.String(), value.Producer,
		value.NotificationType, string(value.Category), string(value.TemplateKind),
		value.Locale, value.Subject, value.Body, value.DataDigest[:],
		value.SubjectDigest[:], value.BodyDigest[:], value.Revision,
		value.InboxRevision, value.CreatedAt, value.ReadAt, value.DismissedAt, value.PurgedAt)
	return err
}
