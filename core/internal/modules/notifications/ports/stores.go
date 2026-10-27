package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

type Binder interface {
	Bind(context.Context, api.Scope) (Stores, error)
}

type Stores interface {
	Effects() EffectStore
	Messages() MessageStore
	Preferences() PreferenceStore
	Templates() TemplateStore
	Providers() ProviderConfigStore
	SMTPSettings() SMTPSettingsStore
	Deliveries() DeliveryStore
	Realtime() RealtimeStore
	Evidence() EvidenceStore
}
