package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) LoadPublishedTemplate(ctx context.Context, id domain.ID,
	locales []string,
) (domain.PublishedTemplate, error) {
	if !store.usable(ctx, false) || !id.Valid() || len(locales) == 0 ||
		len(locales) > 16 {
		return domain.PublishedTemplate{}, ErrInvalid
	}
	var value domain.PublishedTemplate
	var rootID, versionID string
	var variables []byte
	err := store.queryRow(ctx, getPublishedTemplateSQL, id.String(), locales).Scan(
		&rootID, &versionID, &value.NotificationType, &value.Kind, &value.Locale,
		&value.SubjectPattern, &value.BodyPattern, &variables, &value.MandatoryEmail,
		&value.Revision)
	if store.notFound(err) {
		return domain.PublishedTemplate{}, ErrNotFound
	}
	if err == nil {
		value.ID, err = decodeID(rootID)
	}
	if err == nil {
		value.VersionID, err = decodeID(versionID)
	}
	if err == nil {
		err = json.Unmarshal(variables, &value.Variables)
	}
	if err != nil || !value.Valid() {
		return domain.PublishedTemplate{}, fmt.Errorf("decode notification template")
	}
	return value, nil
}
