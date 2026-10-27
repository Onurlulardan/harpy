package domain

import (
	"errors"
	"time"
)

var ErrPreferenceInvalid = errors.New("notification preference is invalid")

type Preference struct {
	RecipientID    ID
	Category       Category
	Channel        Channel
	Enabled        bool
	Explicit       bool
	VerifiedTarget bool
	QuietHours     *QuietHours
	Revision       int64
}

type DeliveryPolicy struct {
	Enabled, Mandatory bool
	NotBefore          time.Time
}

func DefaultPreference(recipient ID, category Category, channel Channel) (Preference, error) {
	if !recipient.Valid() || !category.Valid() || !channel.Valid() {
		return Preference{}, ErrPreferenceInvalid
	}
	enabled := channel == ChannelInApp ||
		channel == ChannelEmail && (category == CategorySecurity ||
			category == CategoryWorkflowAction || category == CategoryBilling)
	if category == CategoryProduct || channel == ChannelTelegram ||
		(channel != ChannelInApp && (category == CategoryDocument || category == CategorySystem)) {
		enabled = false
	}
	return Preference{RecipientID: recipient, Category: category, Channel: channel,
		Enabled: enabled, Revision: 1}, nil
}

func (value Preference) Decide(now time.Time, legalMandatory bool) (DeliveryPolicy, error) {
	if !value.Valid() || now.IsZero() {
		return DeliveryPolicy{}, ErrPreferenceInvalid
	}
	mandatory := value.Channel == ChannelInApp && value.Category != CategoryProduct ||
		value.Category == CategorySecurity && value.Channel == ChannelEmail ||
		legalMandatory && value.Category == CategoryBilling && value.Channel == ChannelEmail
	enabled := value.Enabled || mandatory
	if value.Channel == ChannelTelegram && (!value.Explicit || !value.VerifiedTarget) {
		enabled = false
	}
	notBefore := now.UTC()
	if enabled && !mandatory && value.Channel != ChannelInApp && value.QuietHours != nil {
		var err error
		notBefore, err = value.QuietHours.MaterializeNotBefore(now)
		if err != nil {
			return DeliveryPolicy{}, err
		}
	}
	return DeliveryPolicy{Enabled: enabled, Mandatory: mandatory, NotBefore: notBefore}, nil
}

func (value Preference) Valid() bool {
	return value.RecipientID.Valid() && value.Category.Valid() && value.Channel.Valid() &&
		value.Revision > 0 && (value.QuietHours == nil || value.QuietHours.Valid())
}
