package notifications

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type scopeAuthorizer struct{}

var notificationProducers = map[string]struct{}{
	"analytics": {}, "billing": {}, "catalog": {}, "documents": {},
	"forms": {}, "iam": {}, "notifications": {}, "system": {},
	"work": {}, "workflow": {},
}

var notificationPermissionOperations = map[string]map[string]struct{}{
	application.PermissionInboxRead: {
		OperationInboxList: {}, OperationInboxGet: {}, OperationTicketCreate: {},
		OperationSocketConnect: {}, OperationLegacyConnect: {},
	},
	application.PermissionInboxUpdate: {
		OperationInboxRead: {}, OperationInboxReadAll: {},
	},
	application.PermissionInboxDelete:       {OperationInboxDismiss: {}},
	application.PermissionPreferencesRead:   {OperationPreferencesGet: {}},
	application.PermissionPreferencesWrite:  {OperationPreferencesEdit: {}},
	application.PermissionSend:              {OperationIntentCreate: {}},
	application.PermissionSMTPSettingsRead:  {"tenancy.settings.get": {}},
	application.PermissionSMTPSettingsWrite: {"tenancy.settings.update": {}},
	application.PermissionSMTPSettingsTest:  {"tenancy.settings.test_email": {}},
}

func newScopeAuthorizer() ports.ProducerAuthorizer { return scopeAuthorizer{} }

func (scopeAuthorizer) AuthorizeIntent(ctx context.Context, scope api.Scope,
	producer, notificationType string,
) error {
	if _, found := notificationProducers[producer]; !found ||
		!strings.HasPrefix(notificationType, producer+".") ||
		len(notificationType) == len(producer)+1 {
		return application.ErrForbidden
	}
	return authorizeNotificationScope(ctx, scope, application.PermissionSend)
}

func (scopeAuthorizer) AuthorizeOperation(ctx context.Context, scope api.Scope,
	permission string,
) error {
	return authorizeNotificationScope(ctx, scope, permission)
}

func authorizeNotificationScope(ctx context.Context, abstract api.Scope,
	permission string,
) error {
	scope, ok := abstract.(platformtx.TenantScope)
	if !ok {
		return application.ErrForbidden
	}
	operations, permissionKnown := notificationPermissionOperations[permission]
	_, operationAllowed := operations[scope.OperationID()]
	if !permissionKnown || !operationAllowed || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, scope.Mode()) != nil ||
		platformtx.RequireOwner(scope, "notifications") != nil {
		return application.ErrForbidden
	}
	return nil
}

var _ ports.ProducerAuthorizer = scopeAuthorizer{}
