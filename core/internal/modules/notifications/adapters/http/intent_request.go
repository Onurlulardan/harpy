package notificationhttp

import (
	"encoding/hex"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/google/uuid"
)

type intentDocument struct {
	NotificationType        string                `json:"notification_type"`
	RecipientPrincipalID    uuid.UUID             `json:"recipient_principal_id"`
	TemplateVersionID       uuid.UUID             `json:"template_version_id"`
	ProviderConfigVersionID uuid.UUID             `json:"provider_config_version_id"`
	Locale                  string                `json:"locale"`
	EffectKey               string                `json:"effect_key"`
	Category                string                `json:"category"`
	Channels                []string              `json:"channels"`
	Data                    []intentValueDocument `json:"data"`
	DataDigestSHA256        string                `json:"data_digest_sha256"`
	CorrelationID           uuid.UUID             `json:"correlation_id"`
	CausationID             uuid.UUID             `json:"causation_id"`
}

type intentValueDocument struct {
	Key     string `json:"key"`
	Kind    string `json:"kind"`
	String  string `json:"string,omitempty"`
	Integer int64  `json:"integer,omitempty"`
	Boolean bool   `json:"boolean,omitempty"`
	Time    string `json:"time,omitempty"`
}

func (document intentDocument) request() (api.IntentRequest, error) {
	var digest [32]byte
	decoded, err := hex.DecodeString(document.DataDigestSHA256)
	if err != nil || len(decoded) != len(digest) {
		return api.IntentRequest{}, ErrInvalid
	}
	copy(digest[:], decoded)
	channels := make([]api.Channel, len(document.Channels))
	for index, value := range document.Channels {
		channels[index] = api.Channel(value)
	}
	data := make([]api.TemplateValue, len(document.Data))
	for index, value := range document.Data {
		data[index] = api.TemplateValue{Key: value.Key, Kind: api.ValueKind(value.Kind),
			String: value.String, Integer: value.Integer, Boolean: value.Boolean}
		if value.Time != "" {
			data[index].Time, err = time.Parse(time.RFC3339Nano, value.Time)
			if err != nil {
				return api.IntentRequest{}, ErrInvalid
			}
		}
	}
	return api.IntentRequest{Producer: "notifications",
		NotificationType:        document.NotificationType,
		RecipientPrincipalID:    api.ID(document.RecipientPrincipalID),
		TemplateVersionID:       api.ID(document.TemplateVersionID),
		ProviderConfigVersionID: api.ID(document.ProviderConfigVersionID),
		LocaleInput:             document.Locale, EffectKey: document.EffectKey,
		Category: api.Category(document.Category), Channels: channels, Data: data,
		DataDigest: digest, CorrelationID: api.ID(document.CorrelationID),
		CausationID: api.ID(document.CausationID)}, nil
}
