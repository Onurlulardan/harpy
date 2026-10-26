package authorization

import (
	"testing"

	"github.com/google/uuid"
)

var (
	testActorID      = uuid.MustParse("11000000-0000-4000-8000-000000000001")
	testResourceID   = uuid.MustParse("12000000-0000-4000-8000-000000000001")
	testActionID     = uuid.MustParse("13000000-0000-4000-8000-000000000001")
	testPermissionID = uuid.MustParse("14000000-0000-4000-8000-000000000001")
	testBindingID    = uuid.MustParse("15000000-0000-4000-8000-000000000001")
	testRoleID       = uuid.MustParse("16000000-0000-4000-8000-000000000001")
	testAssignmentID = uuid.MustParse("17000000-0000-4000-8000-000000000001")
	testMembershipID = uuid.MustParse("18000000-0000-4000-8000-000000000001")
	testOrganization = uuid.MustParse("19000000-0000-4000-8000-000000000001")
)

func testRequest(t *testing.T) Request {
	t.Helper()
	request, err := NewRequest("iam.users", "read", 7, 9)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func testSnapshot(
	t *testing.T,
	bindings []BindingFact,
	implications []ImplicationFact,
) Snapshot {
	t.Helper()
	snapshot, err := NewSnapshot(
		&ActorFact{ID: testActorID, Status: "active", TokenVersion: 7,
			AuthorizationVersion: 9},
		&ResourceFact{ID: testResourceID, State: "active", Protected: true},
		&ActionFact{ID: testActionID, State: "active"},
		11, bindings, implications,
	)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func testExactBinding(path SubjectPath) BindingFact {
	fact := BindingFact{
		Path: path, BindingID: testBindingID, BindingState: "active",
		PermissionID: testPermissionID, PermissionMatchKind: "exact",
		PermissionState: "active", PermissionResource: uuidPointer(testResourceID),
		PermissionAction: uuidPointer(testActionID), ResourceState: "active",
		ActionState: "active",
	}
	switch path {
	case SubjectRole:
		fact.AssignmentID = uuidPointer(testAssignmentID)
		fact.AssignmentState = "active"
		fact.RoleID = uuidPointer(testRoleID)
		fact.RoleStatus = "active"
		fact.RoleScopeKind = "tenant"
	case SubjectOrganization:
		fact.MembershipID = uuidPointer(testMembershipID)
		fact.MembershipState = "active"
		fact.OrganizationID = uuidPointer(testOrganization)
		fact.OrganizationStatus = "active"
	}
	return fact
}

func testOrganizationRoleBinding() BindingFact {
	fact := testExactBinding(SubjectRole)
	fact.RoleScopeKind = "organization"
	fact.MembershipID = uuidPointer(testMembershipID)
	fact.MembershipState = "active"
	fact.OrganizationID = uuidPointer(testOrganization)
	fact.OrganizationStatus = "active"
	return fact
}

func testWildcardBinding(kind string) BindingFact {
	fact := testExactBinding(SubjectRole)
	fact.PermissionMatchKind = kind
	fact.PermissionProtected = true
	fact.RoleProtected = true
	fact.PolicyRoleID = uuidPointer(testRoleID)
	fact.PolicyPermissionID = uuidPointer(testPermissionID)
	fact.PolicyBindingID = uuidPointer(testBindingID)
	fact.PolicyState = "active"
	if kind == "resource_all_actions" {
		fact.PermissionAction = nil
		fact.ActionState = ""
	} else {
		fact.PermissionResource = nil
		fact.PermissionAction = nil
		fact.ResourceState = ""
		fact.ActionState = ""
	}
	return fact
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	result := value
	return &result
}
