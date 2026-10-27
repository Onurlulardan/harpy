package postgres

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type scanner interface{ Scan(...any) error }

func scanMessage(row scanner) (domain.Message, error) {
	var value domain.Message
	var id, recipient, template string
	var data, subject, body []byte
	err := row.Scan(&id, &recipient, &template, &value.Producer,
		&value.NotificationType, &value.Category, &value.TemplateKind, &value.Locale,
		&value.Subject, &value.Body, &data, &subject, &body, &value.Revision,
		&value.InboxRevision, &value.CreatedAt, &value.ReadAt, &value.DismissedAt,
		&value.PurgedAt)
	if err != nil {
		return domain.Message{}, err
	}
	value.ID, err = decodeID(id)
	if err == nil {
		value.RecipientID, err = decodeID(recipient)
	}
	if err == nil {
		value.TemplateVersionID, err = decodeID(template)
	}
	if err == nil {
		value.DataDigest, err = decodeDigest(data)
	}
	if err == nil {
		value.SubjectDigest, err = decodeDigest(subject)
	}
	if err == nil {
		value.BodyDigest, err = decodeDigest(body)
	}
	if err != nil || !value.Category.Valid() ||
		(value.TemplateKind != domain.TemplateSystemHTML &&
			value.TemplateKind != domain.TemplatePlainText) || value.Revision < 1 ||
		value.InboxRevision < 1 || value.CreatedAt.IsZero() {
		return domain.Message{}, fmt.Errorf("decode notification message")
	}
	normalizeTimes(&value)
	return value, nil
}

func normalizeTimes(value *domain.Message) {
	value.CreatedAt = value.CreatedAt.UTC()
	for _, pointer := range []**time.Time{&value.ReadAt, &value.DismissedAt, &value.PurgedAt} {
		if *pointer != nil {
			normalized := (*pointer).UTC()
			*pointer = &normalized
		}
	}
}
