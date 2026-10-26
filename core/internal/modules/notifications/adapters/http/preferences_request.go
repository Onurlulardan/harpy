package notificationhttp

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type preferencesDocument struct {
	Updates []preferenceDocument `json:"updates"`
}

type preferenceDocument struct {
	Category         string              `json:"category"`
	Channel          string              `json:"channel"`
	Enabled          bool                `json:"enabled"`
	QuietHours       *quietHoursDocument `json:"quiet_hours"`
	ExpectedRevision int64               `json:"expected_revision"`
}

type quietHoursDocument struct {
	TimeZone    string `json:"time_zone"`
	StartMinute int    `json:"start_minute"`
	EndMinute   int    `json:"end_minute"`
}

func (document preferencesDocument) input(access Access) PreferencesInput {
	updates := make([]application.PreferenceUpdate, len(document.Updates))
	for index, value := range document.Updates {
		var quiet *domain.QuietHours
		if value.QuietHours != nil {
			quiet = &domain.QuietHours{TimeZone: value.QuietHours.TimeZone,
				StartMinute: value.QuietHours.StartMinute, EndMinute: value.QuietHours.EndMinute}
		}
		updates[index] = application.PreferenceUpdate{
			Category: domain.Category(value.Category), Channel: domain.Channel(value.Channel),
			Enabled: value.Enabled, QuietHours: quiet,
			ExpectedRevision: value.ExpectedRevision,
		}
	}
	return PreferencesInput{Access: access, Updates: updates}
}
