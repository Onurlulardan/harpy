package api

import "strings"

func (value IntentRequest) Validate() error {
	if value.Producer != strings.TrimSpace(value.Producer) ||
		value.NotificationType != strings.TrimSpace(value.NotificationType) ||
		value.LocaleInput != strings.TrimSpace(value.LocaleInput) ||
		value.EffectKey != strings.TrimSpace(value.EffectKey) {
		return ErrIntentInvalid
	}
	if value.Producer == "" || len(value.Producer) > 64 ||
		value.NotificationType == "" || len(value.NotificationType) > 96 ||
		!value.RecipientPrincipalID.Valid() || !value.TemplateVersionID.Valid() ||
		value.LocaleInput == "" || len(value.LocaleInput) > 32 ||
		value.EffectKey == "" || len(value.EffectKey) > 512 ||
		!value.Category.Valid() || len(value.Channels) == 0 || len(value.Channels) > 3 ||
		!value.CorrelationID.Valid() || !value.CausationID.Valid() {
		return ErrIntentInvalid
	}
	seen := make(map[Channel]struct{}, len(value.Channels))
	for _, channel := range value.Channels {
		if !channel.Valid() {
			return ErrIntentInvalid
		}
		if _, exists := seen[channel]; exists {
			return ErrIntentInvalid
		}
		seen[channel] = struct{}{}
	}
	digest, ok := CanonicalDataDigest(value.Data)
	if !ok || digest != value.DataDigest {
		return ErrIntentInvalid
	}
	return nil
}
