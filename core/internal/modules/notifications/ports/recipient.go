package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type Recipient struct {
	TenantID, PrincipalID domain.ID
	SupportedLocales      []string
	EmailVerified         bool
	TelegramVerified      bool
	AuthorizationVersion  int64
}

type RecipientResolver interface {
	ResolveRecipient(context.Context, api.Scope, domain.ID) (Recipient, error)
}

type ProducerAuthorizer interface {
	AuthorizeIntent(context.Context, api.Scope, string, string) error
	AuthorizeOperation(context.Context, api.Scope, string) error
}
