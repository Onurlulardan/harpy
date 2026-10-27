package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type PreferenceStore interface {
	LoadPreference(context.Context, domain.ID, domain.Category, domain.Channel) (domain.Preference, bool, error)
	ReadPreferences(context.Context, domain.ID) ([]domain.Preference, error)
	SetPreference(context.Context, domain.Preference, int64) error
}
