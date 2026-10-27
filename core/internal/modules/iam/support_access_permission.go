package iam

type supportAccessTarget struct {
	resource, action, capability string
	platform                     bool
}

var supportAccessTargets = map[string]supportAccessTarget{
	"reporter\x00bug_reports.create\x00support.bug.create":                   {resource: "bug_reports", action: "create"},
	"reporter\x00bug_reports.read_own\x00support.bug.mine":                   {resource: "bug_reports", action: "read_own"},
	"reporter\x00bug_reports.read_own\x00support.bug.read":                   {resource: "bug_reports", action: "read_own"},
	"tenant\x00bug_reports.read\x00support.bug.tenant.list":                  {resource: "bug_reports", action: "read"},
	"tenant\x00bug_reports.read\x00support.bug.tenant.stats":                 {resource: "bug_reports", action: "read"},
	"tenant\x00bug_reports.read\x00support.bug.tenant.read":                  {resource: "bug_reports", action: "read"},
	"tenant\x00bug_reports.manage\x00support.bug.tenant.update":              {resource: "bug_reports", action: "manage"},
	"platform\x00platform.bug_reports.read\x00support.bug.platform.list":     {capability: "platform.bug_reports.read", platform: true},
	"platform\x00platform.bug_reports.read\x00support.bug.platform.stats":    {capability: "platform.bug_reports.read", platform: true},
	"platform\x00platform.bug_reports.read\x00support.bug.platform.read":     {capability: "platform.bug_reports.read", platform: true},
	"platform\x00platform.bug_reports.manage\x00support.bug.platform.update": {capability: "platform.bug_reports.manage", platform: true},
	"platform\x00platform.contact_messages.read\x00support.contact.list":     {capability: "platform.contact_messages.read", platform: true},
	"platform\x00platform.contact_messages.read\x00support.contact.stats":    {capability: "platform.contact_messages.read", platform: true},
	"platform\x00platform.contact_messages.read\x00support.contact.read":     {capability: "platform.contact_messages.read", platform: true},
	"platform\x00platform.contact_messages.manage\x00support.contact.update": {capability: "platform.contact_messages.manage", platform: true},
}

func resolveSupportAccess(scope, permission, operation string) (supportAccessTarget, bool) {
	target, valid := supportAccessTargets[scope+"\x00"+permission+"\x00"+operation]
	return target, valid
}
