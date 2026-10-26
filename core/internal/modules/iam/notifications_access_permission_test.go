package iam

import "testing"

func TestNotificationsAccessTargetsCoverEveryHTTPOperation(t *testing.T) {
	tests := []struct {
		permission, operation, resource, action string
		session, csrf, origin, consumer         bool
	}{
		{"notifications.inbox.read", "notifications.inbox.list",
			"notifications.inbox", "read", false, false, false, false},
		{"notifications.inbox.read", "notifications.inbox.get",
			"notifications.inbox", "read", false, false, false, false},
		{"notifications.inbox.update", "notifications.inbox.mark_read",
			"notifications.inbox", "update", false, false, false, false},
		{"notifications.inbox.update", "notifications.inbox.mark_all_read",
			"notifications.inbox", "update", false, false, false, false},
		{"notifications.inbox.delete", "notifications.inbox.dismiss",
			"notifications.inbox", "delete", false, false, false, false},
		{"notifications.preferences.read", "notifications.preferences.get",
			"notifications.preferences", "read", false, false, false, false},
		{"notifications.preferences.update", "notifications.preferences.update",
			"notifications.preferences", "update", false, false, false, false},
		{"notifications.send", "notifications.intent.create",
			"notifications.intent", "send", false, false, false, false},
		{"notifications.inbox.read", "notifications.realtime.ticket.create",
			"notifications.realtime", "connect", true, true, true, false},
		{"notifications.inbox.read", "notifications.realtime.connect",
			"notifications.realtime", "connect", true, false, true, false},
		{"notifications.inbox.read", "notifications.realtime.legacy_connect",
			"notifications.realtime", "connect", true, false, true, true},
	}
	for _, test := range tests {
		request := NotificationsAccessRequest{
			Permission: test.permission, Operation: test.operation,
			RequireSession: test.session, RequireCSRF: test.csrf,
			RequireHTTPSOrigin: test.origin, RequireSecure: test.origin,
			RequireConsumer: test.consumer,
		}
		target, valid := resolveNotificationsAccess(request)
		if !valid || target.resource != test.resource || target.action != test.action {
			t.Fatalf("%s/%s target = %+v valid=%t",
				test.permission, test.operation, target, valid)
		}
	}
}

func TestNotificationsAccessTargetsRequireExactSecurityFlags(t *testing.T) {
	request := NotificationsAccessRequest{
		Permission: "notifications.inbox.read",
		Operation:  "notifications.realtime.ticket.create", RequireSession: true,
		RequireCSRF: true, RequireHTTPSOrigin: true, RequireSecure: true,
	}
	request.RequireCSRF = false
	if _, valid := resolveNotificationsAccess(request); valid {
		t.Fatal("ticket access accepted weakened CSRF flags")
	}
	request.RequireCSRF = true
	request.RequireConsumer = true
	if _, valid := resolveNotificationsAccess(request); valid {
		t.Fatal("ticket access accepted an unrelated signed-consumer requirement")
	}
}
