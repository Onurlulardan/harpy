package iam

type workflowAccessTarget struct{ resource, action string }

var workflowAccessTargets = map[string]workflowAccessTarget{
	"workflow.definition.create\x00workflow.definition.create":                 {"workflow.definition", "create"},
	"workflow.definition.read\x00workflow.definition.get":                      {"workflow.definition", "read"},
	"workflow.definition.read\x00workflow.definition.list":                     {"workflow.definition", "read"},
	"workflow.definition.update\x00workflow.definition.update_draft":           {"workflow.definition", "update"},
	"workflow.definition.update\x00workflow.definition.rename":                 {"workflow.definition", "update"},
	"workflow.definition.retire\x00workflow.definition.retire":                 {"workflow.definition", "retire"},
	"workflow.publication.publish\x00workflow.publication.publish":             {"workflow.publication", "publish"},
	"workflow.publication.publish\x00workflow.publication_authority.provision": {"workflow.publication", "publish"},
	"workflow.instance.start\x00workflow.runtime.start":                        {"workflow.instance", "start"},
	"workflow.instance.read\x00workflow.runtime.get":                           {"workflow.instance", "read"},
	"workflow.instance.signal\x00workflow.runtime.signal":                      {"workflow.instance", "signal"},
	"workflow.instance.cancel\x00workflow.runtime.cancel":                      {"workflow.instance", "cancel"},
	"workflow.instance.replan\x00workflow.runtime.replan":                      {"workflow.instance", "replan"},
	"workflow.instance.repair\x00workflow.runtime.repair":                      {"workflow.instance", "repair"},
	"workflow.task.read\x00workflow.human_task.get":                            {"workflow.human_tasks", "read"},
	"workflow.task.read\x00workflow.human_task.list":                           {"workflow.human_tasks", "read"},
	"workflow.task.claim\x00workflow.human_task.claim":                         {"workflow.human_tasks", "claim"},
	"workflow.task.decide\x00workflow.human_task.decide":                       {"workflow.human_tasks", "decide"},
	"workflow.task.reassign\x00workflow.human_task.reassign":                   {"workflow.human_tasks", "reassign"},
	"workflow.task.cancel\x00workflow.human_task.cancel":                       {"workflow.human_tasks", "cancel"},
}

var workflowPublicationAuthorityManageTarget = workflowAccessTarget{
	resource: "iam.capability_grant",
	action:   "manage",
}

func resolveWorkflowAccess(permission, operation string) (workflowAccessTarget, bool) {
	target, found := workflowAccessTargets[permission+"\x00"+operation]
	return target, found
}
