package application

import (
	"context"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (application *Application) ReadPreferences(ctx context.Context,
	scope api.Scope,
) ([]domain.Preference, error) {
	stores, err := application.bindOperation(ctx, scope, PermissionPreferencesRead)
	if err != nil {
		return nil, err
	}
	_, recipient, _ := scopeIDs(scope)
	existing, err := stores.Preferences().ReadPreferences(ctx, recipient)
	if err != nil {
		return nil, ErrUnavailable
	}
	byKey := make(map[string]domain.Preference, len(existing))
	for _, preference := range existing {
		if !preference.Valid() || preference.RecipientID != recipient {
			return nil, ErrUnavailable
		}
		byKey[preferenceKey(preference.Category, preference.Channel)] = preference
	}
	result := make([]domain.Preference, 0, 18)
	for _, category := range allCategories() {
		for _, channel := range allChannels() {
			if preference, found := byKey[preferenceKey(category, channel)]; found {
				result = append(result, preference)
				continue
			}
			preference, _ := domain.DefaultPreference(recipient, category, channel)
			result = append(result, preference)
		}
	}
	sort.Slice(result, func(left, right int) bool {
		return preferenceKey(result[left].Category, result[left].Channel) <
			preferenceKey(result[right].Category, result[right].Channel)
	})
	return result, nil
}

func allCategories() []domain.Category {
	return []domain.Category{domain.CategorySecurity, domain.CategoryWorkflowAction,
		domain.CategoryBilling, domain.CategoryDocument, domain.CategorySystem,
		domain.CategoryProduct}
}

func allChannels() []domain.Channel {
	return []domain.Channel{domain.ChannelInApp, domain.ChannelEmail, domain.ChannelTelegram}
}

func preferenceKey(category domain.Category, channel domain.Channel) string {
	return string(category) + ":" + string(channel)
}
