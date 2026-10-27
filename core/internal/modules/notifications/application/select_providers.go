package application

import (
	"context"
	"errors"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type resolvedProviders map[domain.Channel]ports.ProviderConfig

func selectProviders(ctx context.Context, stores ports.Stores,
	request api.IntentRequest,
) (resolvedProviders, error) {
	result := make(resolvedProviders)
	requested, _ := domainID(request.ProviderConfigVersionID)
	for _, selected := range request.Channels {
		if selected == api.ChannelInApp {
			continue
		}
		provider, err := stores.Providers().SelectProvider(ctx, channel(selected), requested)
		if err != nil {
			if errors.Is(err, ports.ErrProviderNotFound) {
				result[channel(selected)] = ports.ProviderConfig{}
				continue
			}
			return nil, ErrUnavailable
		}
		if provider.Kind == "" || !provider.VersionID.Valid() {
			return nil, ErrUnavailable
		}
		result[channel(selected)] = provider
	}
	return result, nil
}

func (values resolvedProviders) digestParts() []string {
	result := make([]string, 0, len(values))
	for channel, provider := range values {
		version := "missing"
		if provider.VersionID.Valid() {
			version = provider.VersionID.String()
		}
		result = append(result, string(channel)+":"+version)
	}
	sort.Strings(result)
	return result
}
