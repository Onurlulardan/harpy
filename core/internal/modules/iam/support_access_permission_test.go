package iam

import "testing"

func TestResolveSupportAccessPinsEveryProtectedHTTPAuthority(t *testing.T) {
	if len(supportAccessTargets) != 15 {
		t.Fatalf("Support protected authority count = %d", len(supportAccessTargets))
	}
	for key, target := range supportAccessTargets {
		if target.platform == (target.capability == "") ||
			(!target.platform && (target.resource != "bug_reports" || target.action == "")) {
			t.Fatalf("Support target %q is invalid: %#v", key, target)
		}
	}
}

func TestResolveSupportAccessRejectsScopePermissionOperationMixing(t *testing.T) {
	for _, input := range [][3]string{
		{"reporter", "bug_reports.read", "support.bug.tenant.read"},
		{"tenant", "bug_reports.read_own", "support.bug.read"},
		{"platform", "platform.bug_reports.read", "support.bug.tenant.list"},
		{"tenant", "bug_reports.manage", "support.bug.tenant.read"},
	} {
		if _, valid := resolveSupportAccess(input[0], input[1], input[2]); valid {
			t.Fatalf("Support authority accepted %q", input)
		}
	}
}
