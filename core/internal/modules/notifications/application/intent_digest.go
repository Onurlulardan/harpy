package application

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func intentDigest(request api.IntentRequest, recipient domain.ID,
	providers resolvedProviders,
) [32]byte {
	hash := sha256.New()
	parts := []string{request.Producer, request.NotificationType,
		request.TemplateVersionID.String(),
		request.LocaleInput, string(request.Category), recipient.String()}
	channels := make([]string, len(request.Channels))
	for index, channel := range request.Channels {
		channels[index] = string(channel)
	}
	sort.Strings(channels)
	parts = append(parts, channels...)
	parts = append(parts, providers.digestParts()...)
	for _, part := range parts {
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(part)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(part))
	}
	_, _ = hash.Write(request.DataDigest[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func renderData(values []api.TemplateValue) (map[string]string, error) {
	result := make(map[string]string, len(values))
	for _, value := range values {
		rendered, ok := value.Render()
		if !ok {
			return nil, ErrInvalid
		}
		result[value.Key] = rendered
	}
	return result, nil
}
