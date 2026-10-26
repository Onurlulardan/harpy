package transaction

func tenantAuthorityGroupMembers(groupID string) ([]string, bool) {
	switch groupID {
	case "analytics":
		return []string{"analytics"}, true
	case "billing_forms":
		return []string{"billing", "forms"}, true
	case "catalog":
		return []string{"catalog"}, true
	case "documents":
		return []string{"documents"}, true
	case "iam_system_resources":
		return []string{"iam", "system_resource_layout"}, true
	case "notifications":
		return []string{"notifications"}, true
	case "tenancy":
		return []string{"tenancy"}, true
	case "work":
		return []string{"work"}, true
	case "workflow":
		return []string{"workflow"}, true
	default:
		return nil, false
	}
}
