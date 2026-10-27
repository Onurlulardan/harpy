package transaction

import "testing"

func TestAuthorityIssuerBindsExactGroupMembersAndGenerations(t *testing.T) {
	generations := map[string]int64{"billing": 4}
	operation, err := NewAuthorityIssuer().Issue(
		"billing.subscription.create", "billing", 3, generations,
	)
	if err != nil || !operation.valid() || !operation.allowsOwner("billing") ||
		operation.allowsOwner("workflow") || operation.mode != Command {
		t.Fatalf("unexpected operation capability: %+v %v", operation, err)
	}
	generations["billing"] = 99
	if operation.domains[0].generation != 4 {
		t.Fatal("operation capability retained caller-owned generation map")
	}
	if _, err := NewAuthorityIssuer().Issue(
		"documents.items.create", "billing", 3, map[string]int64{"forms": 5},
	); err == nil {
		t.Fatal("partial authority group accepted")
	}
}

func TestAuthorityIssuerRequiresExactIAMSystemResourcesGroup(t *testing.T) {
	issuer := NewAuthorityIssuer()
	operation, err := issuer.Issue(
		"iam.users.create", "iam_system_resources", 7,
		map[string]int64{"iam": 3, "system_resource_layout": 5},
	)
	if err != nil || !operation.valid() || !operation.allowsOwner("iam") ||
		!operation.allowsOwner("system_resource_layout") {
		t.Fatalf("IAM system resources authority rejected: %+v %v", operation, err)
	}
	if _, err := issuer.Issue(
		"iam.users.create", "iam", 7, map[string]int64{"iam": 3},
	); err == nil {
		t.Fatal("legacy IAM command group accepted")
	}
}
