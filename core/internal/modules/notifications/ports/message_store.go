package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type MessagePage struct {
	Items      []domain.Message
	NextCursor string
	Unread     int64
}

type MessageFilter struct {
	UnreadOnly bool
	Category   *domain.Category
	Before     string
	Limit      int
}

type MessageStore interface {
	NextInboxRevision(context.Context, domain.ID) (int64, error)
	RecordMessage(context.Context, domain.Message) error
	GetMessage(context.Context, domain.ID, domain.ID) (domain.Message, error)
	ListMessages(context.Context, domain.ID, MessageFilter) (MessagePage, error)
	PersistMessageRevision(context.Context, domain.Message, int64) error
	CompleteInboxRead(context.Context, domain.ID, int64, domain.ID) (int64, error)
	CurrentInboxRevision(context.Context, domain.ID) (int64, error)
}
