package v1bridge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

type payloadValue struct {
	Key   string `json:"key"`
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type bridgePayload struct {
	Version                 int            `json:"version"`
	Producer                string         `json:"producer"`
	NotificationType        string         `json:"notification_type"`
	RecipientPrincipalID    string         `json:"recipient_principal_id"`
	TemplateVersionID       string         `json:"template_version_id"`
	ProviderConfigVersionID string         `json:"provider_config_version_id"`
	LocaleInput             string         `json:"locale_input"`
	Category                string         `json:"category"`
	Channels                []string       `json:"channels"`
	Data                    []payloadValue `json:"data"`
	DataDigest              string         `json:"data_digest"`
	CorrelationID           string         `json:"correlation_id"`
	CausationID             string         `json:"causation_id"`
}

type bridgeIdentity struct {
	Version                 int      `json:"version"`
	Producer                string   `json:"producer"`
	NotificationType        string   `json:"notification_type"`
	RecipientPrincipalID    string   `json:"recipient_principal_id"`
	TemplateVersionID       string   `json:"template_version_id"`
	ProviderConfigVersionID string   `json:"provider_config_version_id"`
	LocaleInput             string   `json:"locale_input"`
	Category                string   `json:"category"`
	Channels                []string `json:"channels"`
	DataDigest              string   `json:"data_digest"`
}

func canonicalPayload(request api.IntentRequest) ([]byte, [32]byte, error) {
	channels := make([]string, len(request.Channels))
	for index, channel := range request.Channels {
		channels[index] = string(channel)
	}
	sort.Strings(channels)
	data := make([]payloadValue, len(request.Data))
	for index, value := range request.Data {
		rendered, ok := value.Render()
		if !ok {
			return nil, [32]byte{}, api.ErrIntentInvalid
		}
		data[index] = payloadValue{Key: value.Key, Kind: string(value.Kind), Value: rendered}
	}
	sort.Slice(data, func(left, right int) bool { return data[left].Key < data[right].Key })
	payload := bridgePayload{Version: 1, Producer: request.Producer,
		NotificationType:        request.NotificationType,
		RecipientPrincipalID:    request.RecipientPrincipalID.String(),
		TemplateVersionID:       request.TemplateVersionID.String(),
		ProviderConfigVersionID: request.ProviderConfigVersionID.String(),
		LocaleInput:             request.LocaleInput, Category: string(request.Category), Channels: channels,
		Data: data, DataDigest: hex.EncodeToString(request.DataDigest[:]),
		CorrelationID: request.CorrelationID.String(), CausationID: request.CausationID.String()}
	document, err := json.Marshal(payload)
	if err != nil {
		return nil, [32]byte{}, ErrStorage
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 64*1024)
	if err != nil {
		return nil, [32]byte{}, api.ErrIntentInvalid
	}
	identity, err := json.Marshal(bridgeIdentity{Version: 1, Producer: payload.Producer,
		NotificationType:        payload.NotificationType,
		RecipientPrincipalID:    payload.RecipientPrincipalID,
		TemplateVersionID:       payload.TemplateVersionID,
		ProviderConfigVersionID: payload.ProviderConfigVersionID,
		LocaleInput:             payload.LocaleInput, Category: payload.Category,
		Channels: payload.Channels, DataDigest: payload.DataDigest})
	if err != nil {
		return nil, [32]byte{}, ErrStorage
	}
	canonicalIdentity, err := canonicaljsonv1.CanonicalizeObject(identity, 16*1024)
	if err != nil {
		return nil, [32]byte{}, api.ErrIntentInvalid
	}
	return canonical, sha256.Sum256(canonicalIdentity), nil
}
