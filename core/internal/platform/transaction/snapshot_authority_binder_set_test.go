package transaction

import "testing"

func TestSnapshotAuthorityIssuerKeepsExactOneToFourBinderOwners(t *testing.T) {
	ownerSets := [][]string{
		{"forms"},
		{"forms", "crm"},
		{"forms", "crm", "workflow"},
		{"forms", "crm", "workflow", "documents"},
	}
	for _, owners := range ownerSets {
		generations := make(map[string]int64, len(owners))
		for _, owner := range owners {
			generations[owner] = 1
		}
		operation, err := NewSnapshotAuthorityIssuer().Issue(
			"restquery.execute", generations,
		)
		if err != nil {
			t.Fatalf("owners %v: %v", owners, err)
		}
		if operation.operationID != "restquery.execute" {
			t.Fatalf("operation ID=%q", operation.operationID)
		}
		for _, owner := range owners {
			if !operation.allowsOwner(owner) {
				t.Fatalf("owner %q missing from %v", owner, owners)
			}
		}
		for _, outside := range []string{
			"forms", "crm", "workflow", "documents", "iam",
		} {
			if !generationsContains(generations, outside) && operation.allowsOwner(outside) {
				t.Fatalf("outside owner %q admitted for %v", outside, owners)
			}
		}
	}
}

func generationsContains(generations map[string]int64, owner string) bool {
	_, found := generations[owner]
	return found
}
