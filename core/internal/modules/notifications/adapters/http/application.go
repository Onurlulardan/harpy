package notificationhttp

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type ListInput struct {
	Access Access
	Filter ports.MessageFilter
}

type MessageInput struct {
	Access Access
	ID     domain.ID
}

type PreferencesInput struct {
	Access  Access
	Updates []application.PreferenceUpdate
}

type IntentInput struct {
	Access  Access
	Request api.IntentRequest
}

type RealtimeInput struct {
	Access           Access
	OfferedProtocols string
	Legacy           bool
}

type Application interface {
	ListMessages(context.Context, ListInput) (ports.MessagePage, error)
	GetMessage(context.Context, MessageInput) (domain.Message, error)
	MarkRead(context.Context, MessageInput) (domain.Message, error)
	MarkAllRead(context.Context, Access) (int64, error)
	DismissMessage(context.Context, MessageInput) (domain.Message, error)
	GetPreferences(context.Context, Access) ([]domain.Preference, error)
	UpdatePreferences(context.Context, PreferencesInput) ([]domain.Preference, error)
	RecordIntent(context.Context, IntentInput) (api.IntentResult, error)
	CreateRealtimeTicket(context.Context, Access) (application.RealtimeTicketResult, error)
	ConsumeRealtimeTicket(context.Context, RealtimeInput) (application.RealtimeConnectionIdentity, error)
}
