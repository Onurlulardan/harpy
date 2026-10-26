package iam

type catalogAccessTarget struct{ resource, action string }

var catalogAccessTargets = map[string]catalogAccessTarget{
	"catalog.read\x00catalog.package.list":                          {"catalog", "read"},
	"catalog.read\x00catalog.package.get":                           {"catalog", "read"},
	"catalog.read\x00catalog.package_version.list":                  {"catalog", "read"},
	"catalog.read\x00catalog.package_version.get":                   {"catalog", "read"},
	"catalog.read\x00catalog.installation.list":                     {"catalog", "read"},
	"catalog.read\x00catalog.installation.get":                      {"catalog", "read"},
	"catalog.read\x00catalog.installation.demo.get":                 {"catalog", "read"},
	"catalog.read\x00catalog.operation.get":                         {"catalog", "read"},
	"catalog.read\x00catalog.menu.get":                              {"catalog", "read"},
	"catalog.install\x00catalog.installation.install":               {"catalog", "install"},
	"catalog.install\x00catalog.operation.cancel":                   {"catalog", "install"},
	"catalog.install\x00catalog.operation.retry":                    {"catalog", "install"},
	"catalog.upgrade\x00catalog.installation.upgrade":               {"catalog", "upgrade"},
	"catalog.disable\x00catalog.installation.disable":               {"catalog", "disable"},
	"catalog.enable\x00catalog.installation.enable":                 {"catalog", "enable"},
	"catalog.restore\x00catalog.installation.restore":               {"catalog", "restore"},
	"catalog.retire\x00catalog.installation.retire":                 {"catalog", "retire"},
	"catalog.purge\x00catalog.installation.purge":                   {"catalog", "purge"},
	"catalog.purge\x00catalog.installation.purge.approve":           {"catalog", "purge"},
	"catalog.demo.manage\x00catalog.installation.demo.removal_plan": {"catalog.demo", "manage"},
	"catalog.demo.manage\x00catalog.installation.demo.install":      {"catalog.demo", "manage"},
	"catalog.demo.manage\x00catalog.installation.demo.remove":       {"catalog.demo", "manage"},
	"catalog.menu.manage\x00catalog.menu.update":                    {"catalog.menu", "manage"},
}

func resolveCatalogAccess(permission, operation string) (catalogAccessTarget, bool) {
	target, found := catalogAccessTargets[permission+"\x00"+operation]
	return target, found
}

func catalogOperationRequiresStepUp(operation string) bool {
	return operation == "catalog.installation.purge" ||
		operation == "catalog.installation.purge.approve"
}
