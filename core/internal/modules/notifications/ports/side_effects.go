package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type RealtimeHint struct {
	MessageID, RecipientID domain.ID
	InboxRevision          int64
}

type RealtimeStore interface {
	AppendHint(context.Context, RealtimeHint) error
}

type RealtimeProjectionStore interface {
	AcquireRealtimeHints(context.Context, int) ([]RealtimeHint, error)
}

type RealtimeProjectionBinder interface {
	BindRealtimeProjection(context.Context, api.Scope) (RealtimeProjectionStore, error)
}

type DeliveryJob struct {
	ID, DeliveryID domain.ID
	Channel        domain.Channel
	EffectKey      string
	RunAt          time.Time
}

type EvidenceStore interface {
	AppendAudit(context.Context, string, domain.ID, domain.ID, int64) error
	EnqueueDelivery(context.Context, DeliveryJob) error
	AppendOperatorAlert(context.Context, string, domain.ID) error
}
