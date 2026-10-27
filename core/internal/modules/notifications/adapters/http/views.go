package notificationhttp

import (
	"encoding/hex"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type messageView struct {
	ID                string     `json:"id"`
	TemplateVersionID string     `json:"template_version_id"`
	Producer          string     `json:"producer"`
	NotificationType  string     `json:"notification_type"`
	Category          string     `json:"category"`
	TemplateKind      string     `json:"template_kind"`
	Locale            string     `json:"locale"`
	Subject           string     `json:"subject"`
	Body              string     `json:"body"`
	DataDigestSHA256  string     `json:"data_digest_sha256"`
	SubjectSHA256     string     `json:"subject_sha256"`
	BodySHA256        string     `json:"body_sha256"`
	Revision          int64      `json:"revision"`
	InboxRevision     int64      `json:"inbox_revision"`
	CreatedAt         time.Time  `json:"created_at"`
	ReadAt            *time.Time `json:"read_at,omitempty"`
	DismissedAt       *time.Time `json:"dismissed_at,omitempty"`
	PurgedAt          *time.Time `json:"purged_at,omitempty"`
}

type messagePageView struct {
	Items      []messageView `json:"items"`
	NextCursor string        `json:"next_cursor,omitempty"`
	Unread     int64         `json:"unread"`
}

func viewMessage(value domain.Message) messageView {
	return messageView{ID: value.ID.String(), TemplateVersionID: value.TemplateVersionID.String(),
		Producer: value.Producer, NotificationType: value.NotificationType,
		Category: string(value.Category), TemplateKind: string(value.TemplateKind),
		Locale: value.Locale, Subject: value.Subject, Body: value.Body,
		DataDigestSHA256: hex.EncodeToString(value.DataDigest[:]),
		SubjectSHA256:    hex.EncodeToString(value.SubjectDigest[:]),
		BodySHA256:       hex.EncodeToString(value.BodyDigest[:]), Revision: value.Revision,
		InboxRevision: value.InboxRevision, CreatedAt: value.CreatedAt,
		ReadAt: value.ReadAt, DismissedAt: value.DismissedAt, PurgedAt: value.PurgedAt}
}

func viewMessagePage(value ports.MessagePage) (messagePageView, error) {
	result := messagePageView{Items: make([]messageView, len(value.Items)), Unread: value.Unread}
	for index, message := range value.Items {
		result.Items[index] = viewMessage(message)
	}
	if value.NextCursor != "" {
		cursor, err := encodeInboxCursor(value.NextCursor)
		if err != nil {
			return messagePageView{}, ErrUnavailable
		}
		result.NextCursor = cursor
	}
	return result, nil
}

type preferenceView struct {
	Category       string              `json:"category"`
	Channel        string              `json:"channel"`
	Enabled        bool                `json:"enabled"`
	Explicit       bool                `json:"explicit"`
	VerifiedTarget bool                `json:"verified_target"`
	QuietHours     *quietHoursDocument `json:"quiet_hours,omitempty"`
	Revision       int64               `json:"revision"`
}

func viewPreferences(values []domain.Preference) []preferenceView {
	result := make([]preferenceView, len(values))
	for index, value := range values {
		result[index] = preferenceView{Category: string(value.Category), Channel: string(value.Channel),
			Enabled: value.Enabled, Explicit: value.Explicit,
			VerifiedTarget: value.VerifiedTarget, Revision: value.Revision}
		if value.QuietHours != nil {
			result[index].QuietHours = &quietHoursDocument{TimeZone: value.QuietHours.TimeZone,
				StartMinute: value.QuietHours.StartMinute, EndMinute: value.QuietHours.EndMinute}
		}
	}
	return result
}
