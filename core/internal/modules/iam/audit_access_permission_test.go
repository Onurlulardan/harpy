package iam

import "testing"

func TestResolveAuditAccessPinsExactPermissionOperations(t *testing.T) {
	for operation := range auditAccessOperations {
		target, valid := resolveAuditAccess("changelog.read", operation)
		if !valid || target.resource != "changelog" || target.action != "read" {
			t.Fatalf("Audit target for %s = %#v/%t", operation, target, valid)
		}
	}
	for _, input := range [][2]string{
		{"changelog.manage", "changelog.list"},
		{"changelog.read", "changelog.delete"},
		{"changelog.read", ""},
	} {
		if _, valid := resolveAuditAccess(input[0], input[1]); valid {
			t.Fatalf("Audit authority accepted %q/%q", input[0], input[1])
		}
	}
}
