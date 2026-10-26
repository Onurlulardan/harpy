package iam

type systemLayoutAccessTarget struct{ permissionAction string }

var systemLayoutAccessTargets = map[string]systemLayoutAccessTarget{
	"read\x00system_resource_layout.get":         {permissionAction: "read"},
	"read\x00system_resource_layout.list":        {permissionAction: "read"},
	"read\x00system_resource_layout.resolve":     {permissionAction: "read"},
	"update\x00system_resource_layout.create":    {permissionAction: "update"},
	"update\x00system_resource_layout.update":    {permissionAction: "update"},
	"update\x00system_resource_layout.delete":    {permissionAction: "update"},
	"update\x00system_resource_layout.publish":   {permissionAction: "update"},
	"read\x00iam.identity.user.list":             {permissionAction: "read"},
	"read\x00iam.identity.user.read":             {permissionAction: "read"},
	"create\x00iam.identity.user.create":         {permissionAction: "create"},
	"update\x00iam.identity.user.update":         {permissionAction: "update"},
	"delete\x00iam.identity.user.delete":         {permissionAction: "delete"},
	"read\x00iam.identity.role.list":             {permissionAction: "read"},
	"read\x00iam.identity.role.read":             {permissionAction: "read"},
	"create\x00iam.identity.role.create":         {permissionAction: "create"},
	"update\x00iam.identity.role.update":         {permissionAction: "update"},
	"delete\x00iam.identity.role.retire":         {permissionAction: "delete"},
	"read\x00iam.identity.organization.list":     {permissionAction: "read"},
	"read\x00iam.identity.organization.read":     {permissionAction: "read"},
	"create\x00iam.identity.organization.create": {permissionAction: "create"},
	"update\x00iam.identity.organization.update": {permissionAction: "update"},
	"delete\x00iam.identity.organization.delete": {permissionAction: "delete"},
}

func resolveSystemLayoutAccess(action, operation string) (systemLayoutAccessTarget, bool) {
	target, found := systemLayoutAccessTargets[action+"\x00"+operation]
	return target, found
}
