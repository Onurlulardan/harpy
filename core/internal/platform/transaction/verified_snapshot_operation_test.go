package transaction

import (
	"testing"

	"github.com/google/uuid"
)

func TestSnapshotAuthorityIssuerOwnsCanonicalExactDomains(t *testing.T) {
	generations := map[string]int64{"forms": 7, "analytics": 3}
	operation, err := NewSnapshotAuthorityIssuer().Issue(
		"restquery.execute", generations,
	)
	if err != nil || !operation.valid() || operation.mode != SnapshotQuery ||
		len(operation.groups) != 0 || len(operation.participants) != 0 ||
		!operation.allowsOwner("analytics") || !operation.allowsOwner("forms") ||
		operation.allowsOwner("iam") || operation.domains[0].name != "analytics" ||
		operation.domains[1].name != "forms" {
		t.Fatalf("unexpected snapshot capability: %+v %v", operation, err)
	}
	generations["forms"] = 99
	if operation.domains[1].generation != 7 {
		t.Fatal("snapshot capability retained caller generation map")
	}
}

func TestSnapshotAuthorityIssuerRejectsInvalidOwnerSets(t *testing.T) {
	issuer := NewSnapshotAuthorityIssuer()
	cases := []map[string]int64{
		nil,
		{"INVALID": 1},
		{"forms": 0},
		{"analytics": 1, "billing": 1, "catalog": 1, "documents": 1, "forms": 1},
	}
	for index, generations := range cases {
		if _, err := issuer.Issue("restquery.execute", generations); err == nil {
			t.Fatalf("invalid owner set %d was accepted", index)
		}
	}
	if _, err := (SnapshotAuthorityIssuer{}).Issue(
		"restquery.execute", map[string]int64{"forms": 1},
	); err == nil {
		t.Fatal("zero issuer was accepted")
	}
}

func TestScopeRequestRequiresExactOperationMode(t *testing.T) {
	actor, _ := NewActorIssuer().Issue(
		uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		uuid.MustParse("22222222-2222-4222-8222-222222222222"),
	)
	command, _ := NewAuthorityIssuer().Issue(
		"iam.users.create", "iam_system_resources", 1,
		map[string]int64{"iam": 1, "system_resource_layout": 1},
	)
	query, _ := NewSnapshotAuthorityIssuer().Issue(
		"restquery.execute", map[string]int64{"forms": 1},
	)
	cases := []struct {
		name      string
		operation VerifiedOperation
		mode      ScopeMode
		valid     bool
	}{
		{"command", command, Command, true},
		{"command as query", command, SnapshotQuery, false},
		{"query", query, SnapshotQuery, true},
		{"query as command", query, Command, false},
	}
	for _, test := range cases {
		request := ScopeRequest{Actor: actor, Operation: test.operation, Mode: test.mode}
		if request.valid() != test.valid {
			t.Fatalf("%s valid=%t operation=%+v", test.name, request.valid(), test.operation)
		}
	}
}

func TestSnapshotOperationTamperFailsClosed(t *testing.T) {
	operation, err := NewSnapshotAuthorityIssuer().Issue(
		"restquery.execute", map[string]int64{"crm": 1, "forms": 1},
	)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := operation
	duplicate.domains = append(
		append([]domainExpectation{}, operation.domains...), operation.domains[1],
	)
	wrongMode := operation
	wrongMode.mode = Command
	withGroup := operation
	withGroup.groups = []groupExpectation{{id: "billing_forms", generation: 1}}
	if duplicate.valid() || wrongMode.valid() || withGroup.valid() {
		t.Fatal("tampered snapshot operation remained valid")
	}
}
