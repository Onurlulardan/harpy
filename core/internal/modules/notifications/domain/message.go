package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrMessageInvalid  = errors.New("notification message is invalid")
	ErrMessageConflict = errors.New("notification message revision conflicts")
)

type Message struct {
	ID, RecipientID, TemplateVersionID ID
	Producer, NotificationType         string
	Category                           Category
	TemplateKind                       TemplateKind
	Locale, Subject, Body              string
	DataDigest, SubjectDigest          [32]byte
	BodyDigest                         [32]byte
	Revision, InboxRevision            int64
	CreatedAt                          time.Time
	ReadAt, DismissedAt, PurgedAt      *time.Time
}

func NewMessage(value Message) (Message, error) {
	value.Producer = strings.TrimSpace(value.Producer)
	value.NotificationType = strings.TrimSpace(value.NotificationType)
	value.Locale = strings.TrimSpace(value.Locale)
	if !value.ID.Valid() || !value.RecipientID.Valid() ||
		!value.TemplateVersionID.Valid() || !value.Category.Valid() ||
		(value.TemplateKind != TemplateSystemHTML && value.TemplateKind != TemplatePlainText) ||
		value.Producer == "" || len(value.Producer) > 64 ||
		value.NotificationType == "" || len(value.NotificationType) > 96 ||
		value.Locale == "" || len(value.Locale) > 16 ||
		value.Subject == "" || len(value.Subject) > 512 ||
		value.Body == "" || len(value.Body) > 64*1024 ||
		value.Revision != 1 || value.InboxRevision < 1 || value.CreatedAt.IsZero() {
		return Message{}, ErrMessageInvalid
	}
	value.CreatedAt = value.CreatedAt.UTC()
	return value, nil
}

func (value Message) MarkRead(expected int64, at time.Time) (Message, error) {
	if expected != value.Revision || at.IsZero() || value.DismissedAt != nil {
		return Message{}, ErrMessageConflict
	}
	if value.ReadAt != nil {
		return value, nil
	}
	readAt := at.UTC()
	value.ReadAt = &readAt
	value.Revision++
	return value, nil
}

func (value Message) Dismiss(expected int64, at time.Time) (Message, error) {
	if expected != value.Revision || at.IsZero() {
		return Message{}, ErrMessageConflict
	}
	if value.DismissedAt != nil {
		return value, nil
	}
	dismissedAt := at.UTC()
	value.DismissedAt = &dismissedAt
	value.Revision++
	return value, nil
}

func (value Message) PurgeContent(at time.Time) (Message, error) {
	if at.IsZero() || value.PurgedAt != nil {
		return Message{}, ErrMessageConflict
	}
	purgedAt := at.UTC()
	value.Subject, value.Body = "", ""
	value.PurgedAt = &purgedAt
	value.Revision++
	return value, nil
}
