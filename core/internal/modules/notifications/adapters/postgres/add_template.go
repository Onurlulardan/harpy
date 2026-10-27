package postgres

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) AddTemplate(ctx context.Context,
	draft ports.TemplateDraft,
) error {
	if !store.usable(ctx, true) || !draft.ID.Valid() || draft.NotificationType == "" ||
		draft.Revision != 1 {
		return ErrInvalid
	}
	_, err := store.execute(ctx, insertTemplateSQL, draft.ID.String(),
		draft.NotificationType, draft.Revision)
	return err
}

func (store *Store) StoreTemplateVersion(ctx context.Context,
	version ports.TemplateVersion,
) error {
	value := version.Template
	if !store.usable(ctx, true) || !value.Valid() || version.Digest == [32]byte{} {
		return ErrInvalid
	}
	variables, err := json.Marshal(value.Variables)
	if err != nil {
		return ErrInvalid
	}
	defer clear(variables)
	_, err = store.execute(ctx, insertTemplateVersionSQL, value.VersionID.String(),
		value.ID.String(), value.Revision, value.Locale, string(value.Kind),
		value.SubjectPattern, value.BodyPattern, variables, value.MandatoryEmail,
		version.Digest[:])
	return err
}
