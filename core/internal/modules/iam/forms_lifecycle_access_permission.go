package iam

type formsLifecycleAccessTarget struct{ resource, action string }

var formsLifecycleAccessTargets = map[string]formsLifecycleAccessTarget{
	"forms.definition.create\x00forms.definition.create":           {"forms.definition", "create"},
	"forms.definition.read\x00forms.definition.get":                {"forms.definition", "read"},
	"forms.definition.read\x00forms.definition.list":               {"forms.definition", "read"},
	"forms.definition.update\x00forms.definition.update_draft":     {"forms.definition", "update"},
	"forms.definition.update\x00forms.definition.create_version":   {"forms.definition", "update"},
	"forms.definition.retire\x00forms.definition.retire":           {"forms.definition", "retire"},
	"forms.publication.create\x00forms.publication.create":         {"forms.publication", "create"},
	"forms.publication.read\x00forms.publication.get":              {"forms.publication", "read"},
	"forms.publication.validate\x00forms.publication.validate":     {"forms.publication", "validate"},
	"forms.publication.plan\x00forms.publication.plan":             {"forms.publication", "plan"},
	"forms.publication.activate\x00forms.publication.activate":     {"forms.publication", "activate"},
	"forms.publication.deactivate\x00forms.publication.deactivate": {"forms.publication", "deactivate"},
}

func resolveFormsLifecycleAccess(
	permission, operation string,
) (formsLifecycleAccessTarget, bool) {
	target, found := formsLifecycleAccessTargets[permission+"\x00"+operation]
	return target, found
}
