package api

import "errors"

var (
	ErrIntentInvalid           = errors.New("NOTIFICATION_INTENT_INVALID")
	ErrEffectKeyReused         = errors.New("NOTIFICATION_EFFECT_KEY_REUSED")
	ErrRecipientNotFound       = errors.New("NOTIFICATION_RECIPIENT_NOT_FOUND")
	ErrTemplateNotFound        = errors.New("NOTIFICATION_TEMPLATE_NOT_FOUND")
	ErrSMTPSettingsInvalid     = errors.New("NOTIFICATION_SMTP_SETTINGS_INVALID")
	ErrSMTPSettingsNotFound    = errors.New("NOTIFICATION_SMTP_SETTINGS_NOT_FOUND")
	ErrSMTPSettingsConflict    = errors.New("NOTIFICATION_SMTP_SETTINGS_CONFLICT")
	ErrSMTPSettingsForbidden   = errors.New("NOTIFICATION_SMTP_SETTINGS_FORBIDDEN")
	ErrSMTPSettingsUnavailable = errors.New("NOTIFICATION_SMTP_SETTINGS_UNAVAILABLE")
	ErrForbidden               = errors.New("NOTIFICATION_FORBIDDEN")
	ErrUnavailable             = errors.New("NOTIFICATION_UNAVAILABLE")
)
