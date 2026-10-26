package iam

type notificationsAccessTarget struct {
	resource, action      string
	session, csrf, origin bool
	signedConsumer        bool
}

var notificationsAccessTargets = map[string]notificationsAccessTarget{
	"notifications.inbox.read\x00notifications.inbox.list":                 {"notifications.inbox", "read", false, false, false, false},
	"notifications.inbox.read\x00notifications.inbox.get":                  {"notifications.inbox", "read", false, false, false, false},
	"notifications.inbox.update\x00notifications.inbox.mark_read":          {"notifications.inbox", "update", false, false, false, false},
	"notifications.inbox.update\x00notifications.inbox.mark_all_read":      {"notifications.inbox", "update", false, false, false, false},
	"notifications.inbox.delete\x00notifications.inbox.dismiss":            {"notifications.inbox", "delete", false, false, false, false},
	"notifications.preferences.read\x00notifications.preferences.get":      {"notifications.preferences", "read", false, false, false, false},
	"notifications.preferences.update\x00notifications.preferences.update": {"notifications.preferences", "update", false, false, false, false},
	"notifications.send\x00notifications.intent.create":                    {"notifications.intent", "send", false, false, false, false},
	"notifications.inbox.read\x00notifications.realtime.ticket.create":     {"notifications.realtime", "connect", true, true, true, false},
	"notifications.inbox.read\x00notifications.realtime.connect":           {"notifications.realtime", "connect", true, false, true, false},
	"notifications.inbox.read\x00notifications.realtime.legacy_connect":    {"notifications.realtime", "connect", true, false, true, true},
}

func resolveNotificationsAccess(request NotificationsAccessRequest) (
	notificationsAccessTarget, bool,
) {
	target, found := notificationsAccessTargets[request.Permission+"\x00"+request.Operation]
	validFlags := request.RequireSession == target.session &&
		request.RequireCSRF == target.csrf &&
		request.RequireHTTPSOrigin == target.origin &&
		request.RequireSecure == target.origin &&
		request.RequireConsumer == target.signedConsumer
	return target, found && validFlags
}
