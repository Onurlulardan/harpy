package notificationhttp

const (
	OperationInboxList       = "notifications.inbox.list"
	OperationInboxGet        = "notifications.inbox.get"
	OperationInboxRead       = "notifications.inbox.mark_read"
	OperationInboxReadAll    = "notifications.inbox.mark_all_read"
	OperationInboxDismiss    = "notifications.inbox.dismiss"
	OperationPreferencesGet  = "notifications.preferences.get"
	OperationPreferencesEdit = "notifications.preferences.update"
	OperationIntentCreate    = "notifications.intent.create"
	OperationTicketCreate    = "notifications.realtime.ticket.create"
	OperationSocketConnect   = "notifications.realtime.connect"
	OperationLegacyConnect   = "notifications.realtime.legacy_connect"
)

func operationSecurity(operation string) (session, csrf, origin bool) {
	switch operation {
	case OperationTicketCreate:
		return true, true, true
	case OperationSocketConnect, OperationLegacyConnect:
		return true, false, true
	default:
		return false, false, false
	}
}

func validPermissionOperation(permission Permission, operation string) bool {
	switch permission {
	case PermissionInboxRead:
		return operation == OperationInboxList || operation == OperationInboxGet ||
			operation == OperationTicketCreate || operation == OperationSocketConnect ||
			operation == OperationLegacyConnect
	case PermissionInboxUpdate:
		return operation == OperationInboxRead || operation == OperationInboxReadAll
	case PermissionInboxDelete:
		return operation == OperationInboxDismiss
	case PermissionPreferencesRead:
		return operation == OperationPreferencesGet
	case PermissionPreferencesEdit:
		return operation == OperationPreferencesEdit
	case PermissionSend:
		return operation == OperationIntentCreate
	default:
		return false
	}
}
