package application

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestSetPreferencesRejectsMandatoryOptOutAndTelegramWithoutVerification(t *testing.T) {
	application, _, scope, _ := newIntentFixture(t)
	_, err := application.SetPreferences(t.Context(), scope, []PreferenceUpdate{{
		Category: domain.CategorySecurity, Channel: domain.ChannelEmail,
		Enabled: false, ExpectedRevision: 0}})
	if err != ErrInvalid {
		t.Fatalf("mandatory opt-out error = %v", err)
	}
	_, err = application.SetPreferences(t.Context(), scope, []PreferenceUpdate{{
		Category: domain.CategoryProduct, Channel: domain.ChannelTelegram,
		Enabled: true, ExpectedRevision: 0}})
	if err != ErrInvalid {
		t.Fatalf("unverified Telegram error = %v", err)
	}
}

func TestSetPreferencesUsesRevisionCAS(t *testing.T) {
	application, _, scope, _ := newIntentFixture(t)
	update := PreferenceUpdate{Category: domain.CategoryProduct,
		Channel: domain.ChannelEmail, Enabled: true, ExpectedRevision: 0}
	values, err := application.SetPreferences(t.Context(), scope, []PreferenceUpdate{update})
	if err != nil || values[0].Revision != 1 {
		t.Fatalf("SetPreferences() = %+v, %v", values, err)
	}
	if _, err = application.SetPreferences(t.Context(), scope,
		[]PreferenceUpdate{update}); err != ErrConflict {
		t.Fatalf("stale preference error = %v", err)
	}
}
