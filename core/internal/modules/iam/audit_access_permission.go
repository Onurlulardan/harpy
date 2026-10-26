package iam

type auditAccessTarget struct{ resource, action string }

var auditAccessOperations = map[string]struct{}{
	"changelog.list":                {},
	"changelog.stats":               {},
	"changelog.get":                 {},
	"changelog.entity_history":      {},
	"changelog.form_record_history": {},
}

func resolveAuditAccess(permission, operation string) (auditAccessTarget, bool) {
	if permission != "changelog.read" {
		return auditAccessTarget{}, false
	}
	_, valid := auditAccessOperations[operation]
	return auditAccessTarget{resource: "changelog", action: "read"}, valid
}
