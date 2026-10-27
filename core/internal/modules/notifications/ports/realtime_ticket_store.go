package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type RealtimeTicketStore interface {
	InsertTicket(context.Context, domain.Ticket) error
	ConsumeTicket(context.Context, domain.TicketConsume) (domain.Ticket, error)
}

type RealtimeTicketBinder interface {
	BindRealtime(context.Context, api.Scope) (RealtimeTicketStore, error)
}

type LegacyConnectionHit struct {
	TenantID, UserID       domain.ID
	ConsumerManifestDigest [32]byte
}

type LegacyConnectionStore interface {
	InsertLegacyConnectionHit(context.Context, LegacyConnectionHit) error
}

type LegacyConnectionBinder interface {
	BindLegacyConnection(context.Context, api.Scope) (LegacyConnectionStore, error)
}
