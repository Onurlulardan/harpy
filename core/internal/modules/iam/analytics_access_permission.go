package iam

type analyticsAccessTarget struct{ resource, action string }

var analyticsAccessTargets = map[string]analyticsAccessTarget{
	"analytics.query.create":                   {"analytics.query", "create"},
	"analytics.query.read":                     {"analytics.query", "read"},
	"analytics.query.update":                   {"analytics.query", "update"},
	"analytics.query.delete":                   {"analytics.query", "delete"},
	"analytics.query.publish":                  {"analytics.query", "publish"},
	"analytics.query.execute":                  {"analytics.query", "execute"},
	"analytics.dashboard.create":               {"analytics.dashboard", "create"},
	"analytics.dashboard.read":                 {"analytics.dashboard", "read"},
	"analytics.dashboard.update":               {"analytics.dashboard", "update"},
	"analytics.dashboard.delete":               {"analytics.dashboard", "delete"},
	"analytics.dashboard.publish":              {"analytics.dashboard", "publish"},
	"analytics.dashboard.export":               {"analytics.dashboard", "export"},
	"analytics.publication.read":               {"analytics.publication", "read"},
	"analytics.publication.revoke":             {"analytics.publication", "revoke"},
	"analytics.publication.publish_public":     {"analytics.publication", "publish_public"},
	"analytics.publication.approve_public":     {"analytics.publication", "approve_public"},
	"analytics.publication.publish_capability": {"analytics.publication", "publish_capability"},
	"analytics.publication.approve_capability": {"analytics.publication", "approve_capability"},
}

func resolveAnalyticsAccess(permission, operation string) (analyticsAccessTarget, bool) {
	target, found := analyticsAccessTargets[permission]
	return target, found && analyticsPermissionOperation(permission, operation)
}

func analyticsPermissionOperation(permission, operation string) bool {
	switch permission {
	case "analytics.query.read":
		return operation == "analytics.query.get" || operation == "analytics.query.list"
	case "analytics.dashboard.read":
		return operation == "analytics.dashboard.read" || operation == "analytics.dashboard.get" ||
			operation == "analytics.dashboard.list" ||
			operation == "analytics.dashboard.set_default"
	case "analytics.dashboard.update":
		return operation == "analytics.dashboard.update" ||
			operation == "analytics.dashboard.grant_acl" || operation == "analytics.dashboard.revoke_acl"
	case "analytics.dashboard.export":
		return operation == "analytics.dashboard.export.create" ||
			operation == "analytics.dashboard.export.get" ||
			operation == "analytics.dashboard.export.delete" ||
			operation == "analytics.dashboard.export.content"
	case "analytics.publication.read":
		return operation == "analytics.publication.create" || operation == "analytics.publication.get" ||
			operation == "analytics.publication.list" || operation == "analytics.publication.submit" ||
			operation == "analytics.publication.approve" || operation == "analytics.publication.activate"
	case "analytics.publication.publish_public",
		"analytics.publication.publish_capability":
		return operation == "analytics.publication.create"
	default:
		return permission == operation ||
			(permission == "analytics.query.publish" && operation == "analytics.query.revalidate")
	}
}

func analyticsOperationRequiresStepUp(operation string) bool {
	return operation == "analytics.publication.submit" ||
		operation == "analytics.publication.approve" ||
		operation == "analytics.publication.activate"
}
