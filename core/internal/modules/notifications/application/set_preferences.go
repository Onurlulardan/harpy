package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type PreferenceUpdate struct {
	Category         domain.Category
	Channel          domain.Channel
	Enabled          bool
	QuietHours       *domain.QuietHours
	ExpectedRevision int64
}

func (application *Application) SetPreferences(ctx context.Context,
	scope api.Scope, updates []PreferenceUpdate,
) ([]domain.Preference, error) {
	if len(updates) == 0 || len(updates) > 18 {
		return nil, ErrInvalid
	}
	stores, err := application.bindOperation(ctx, scope, PermissionPreferencesWrite)
	if err != nil {
		return nil, err
	}
	_, recipientID, _ := scopeIDs(scope)
	recipient, err := application.recipients.ResolveRecipient(ctx, scope, recipientID)
	if err != nil || recipient.PrincipalID != recipientID {
		return nil, ErrUnavailable
	}
	seen := make(map[string]struct{}, len(updates))
	result := make([]domain.Preference, 0, len(updates))
	for _, update := range updates {
		key := preferenceKey(update.Category, update.Channel)
		if _, exists := seen[key]; exists || !validPreferenceUpdate(update) {
			return nil, ErrInvalid
		}
		seen[key] = struct{}{}
		current, found, readErr := stores.Preferences().LoadPreference(
			ctx, recipientID, update.Category, update.Channel)
		if readErr != nil {
			return nil, ErrUnavailable
		}
		if !found {
			current, _ = domain.DefaultPreference(recipientID, update.Category, update.Channel)
			current.Revision = 0
		}
		if current.Revision != update.ExpectedRevision {
			return nil, ErrConflict
		}
		current.Enabled, current.Explicit = update.Enabled, true
		current.QuietHours, current.Revision = update.QuietHours, current.Revision+1
		if update.Channel == domain.ChannelEmail {
			current.VerifiedTarget = recipient.EmailVerified
		} else if update.Channel == domain.ChannelTelegram {
			current.VerifiedTarget = recipient.TelegramVerified
		}
		if !current.Valid() || current.Enabled && current.Channel == domain.ChannelTelegram &&
			!current.VerifiedTarget {
			return nil, ErrInvalid
		}
		if err = stores.Preferences().SetPreference(ctx, current, update.ExpectedRevision); err != nil {
			return nil, ErrConflict
		}
		result = append(result, current)
	}
	return result, nil
}

func validPreferenceUpdate(value PreferenceUpdate) bool {
	if !value.Category.Valid() || !value.Channel.Valid() || value.ExpectedRevision < 0 ||
		value.QuietHours != nil && (value.Channel == domain.ChannelInApp ||
			!value.QuietHours.Valid()) {
		return false
	}
	mandatory := value.Category != domain.CategoryProduct && value.Channel == domain.ChannelInApp ||
		value.Category == domain.CategorySecurity && value.Channel == domain.ChannelEmail
	return !mandatory || value.Enabled
}
