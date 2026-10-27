package notificationhttp

import (
	"context"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type RealtimeHint struct {
	TenantID, UserID, MessageID domain.ID
	InboxRevision               int64
}

type RealtimeTransport interface {
	Serve(http.ResponseWriter, *http.Request, application.RealtimeConnectionIdentity) error
	Broadcast(context.Context, RealtimeHint) error
}
