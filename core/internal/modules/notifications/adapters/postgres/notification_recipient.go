package postgres

import "github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"

type notificationRecipient struct {
	principalID          domain.ID
	authorizationVersion int64
	supportedLocales     []string
	emailTarget          *string
	emailVerified        bool
	telegramTarget       *string
	telegramVerified     bool
}

func scanNotificationRecipient(source row) (notificationRecipient, error) {
	var value notificationRecipient
	err := source.Scan(&value.principalID, &value.authorizationVersion,
		&value.supportedLocales, &value.emailTarget, &value.emailVerified,
		&value.telegramTarget, &value.telegramVerified)
	if err != nil || !value.principalID.Valid() || value.authorizationVersion < 1 ||
		len(value.supportedLocales) == 0 ||
		value.emailVerified != (value.emailTarget != nil) ||
		value.telegramVerified != (value.telegramTarget != nil) {
		return notificationRecipient{}, ErrInvalid
	}
	return value, nil
}

func (value notificationRecipient) target(channel domain.Channel) ([]byte, error) {
	var target *string
	var verified bool
	switch channel {
	case domain.ChannelEmail:
		target, verified = value.emailTarget, value.emailVerified
	case domain.ChannelTelegram:
		target, verified = value.telegramTarget, value.telegramVerified
	default:
		return nil, ErrInvalid
	}
	if !verified || target == nil || *target == "" {
		return nil, ErrInvalid
	}
	return []byte(*target), nil
}
