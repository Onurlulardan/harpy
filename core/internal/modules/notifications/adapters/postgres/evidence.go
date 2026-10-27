package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) AppendAudit(ctx context.Context, action string,
	resource, actor domain.ID, revision int64,
) error {
	if !store.usable(ctx, true) || action == "" || !resource.Valid() ||
		!actor.Valid() || revision < 1 {
		return ErrInvalid
	}
	_, err := store.execute(ctx, appendAuditSQL, action, resource.String(),
		actor.String(), revision)
	return err
}

func (store *Store) AppendHint(ctx context.Context, hint ports.RealtimeHint) error {
	if !store.usable(ctx, true) || !hint.MessageID.Valid() ||
		!hint.RecipientID.Valid() || hint.InboxRevision < 1 {
		return ErrInvalid
	}
	_, err := store.execute(ctx, appendRealtimeHintSQL, hint.MessageID.String(),
		hint.RecipientID.String(), hint.InboxRevision)
	return err
}

func (store *Store) AppendOperatorAlert(ctx context.Context, code string,
	resource domain.ID,
) error {
	if !store.usable(ctx, true) || code == "" || !resource.Valid() {
		return ErrInvalid
	}
	_, err := store.execute(ctx, appendOperatorAlertSQL, code, resource.String())
	return err
}
