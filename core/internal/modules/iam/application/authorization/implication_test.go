package authorization

import (
	"testing"

	"github.com/google/uuid"
)

func TestEvaluateDoesNotGrantFormsParentImplication(t *testing.T) {
	sourceResource := uuid.MustParse("21000000-0000-4000-8000-000000000001")
	sourceAction := uuid.MustParse("22000000-0000-4000-8000-000000000001")
	sourcePermission := uuid.MustParse("23000000-0000-4000-8000-000000000001")
	binding := testExactBinding(SubjectRole)
	binding.PermissionID = sourcePermission
	binding.PermissionResource = uuidPointer(sourceResource)
	binding.PermissionAction = uuidPointer(sourceAction)
	edge := ImplicationFact{
		ID:   uuid.MustParse("24000000-0000-4000-8000-000000000001"),
		Kind: "v1_form_parent_v1", State: "active",
		SourcePermissionID: sourcePermission, SourcePermissionState: "active",
		SourceResourceID: sourceResource, SourceActionID: sourceAction,
		SourceResourceState: "active", SourceActionState: "active",
		ImpliedPermissionID: testPermissionID, ImpliedPermissionState: "active",
		ImpliedResourceID: testResourceID, ImpliedActionID: testActionID,
		ImpliedResourceState: "active", ImpliedActionState: "active",
	}
	decision := evaluate(
		testRequest(t), testSnapshot(t, []BindingFact{binding}, []ImplicationFact{edge}),
	)
	if decision.Allowed() || decision.Reason() != ReasonNoEffectivePermission {
		t.Fatalf("forms parent implication granted access: %t/%s",
			decision.Allowed(), decision.Reason())
	}
	edge.State = "retired"
	decision = evaluate(
		testRequest(t), testSnapshot(t, []BindingFact{binding}, []ImplicationFact{edge}),
	)
	if decision.Allowed() {
		t.Fatal("retired implication allowed access")
	}
}

func TestEvaluateDoesNotChainOrInventHierarchy(t *testing.T) {
	binding := testExactBinding(SubjectUser)
	edge := ImplicationFact{
		ID:   uuid.MustParse("25000000-0000-4000-8000-000000000001"),
		Kind: "v1_form_parent_v1", State: "active",
		SourcePermissionID:    uuid.MustParse("26000000-0000-4000-8000-000000000001"),
		SourcePermissionState: "active",
		SourceResourceID:      testResourceID, SourceActionID: testActionID,
		SourceResourceState: "active", SourceActionState: "active",
		ImpliedPermissionID: testPermissionID, ImpliedPermissionState: "active",
		ImpliedResourceID: testResourceID, ImpliedActionID: testActionID,
		ImpliedResourceState: "active", ImpliedActionState: "active",
	}
	decision := evaluate(
		testRequest(t), testSnapshot(t, []BindingFact{binding}, []ImplicationFact{edge}),
	)
	if decision.Reason() != ReasonDirectUserExact {
		t.Fatalf("unrelated edge changed direct decision to %s", decision.Reason())
	}
}

func TestEvaluateDoesNotGrantFormsRecordOperation(t *testing.T) {
	for _, path := range []SubjectPath{SubjectUser, SubjectRole, SubjectOrganization} {
		t.Run(string(path), func(t *testing.T) {
			sourceResource := uuid.MustParse("27000000-0000-4000-8000-000000000001")
			sourceAction := uuid.MustParse("27000000-0000-4000-8000-000000000002")
			sourcePermission := uuid.MustParse("27000000-0000-4000-8000-000000000003")
			binding := testExactBinding(path)
			binding.PermissionID = sourcePermission
			binding.PermissionResource = uuidPointer(sourceResource)
			binding.PermissionAction = uuidPointer(sourceAction)
			edge := ImplicationFact{
				ID:   uuid.MustParse("27000000-0000-4000-8000-000000000004"),
				Kind: "forms_record_operation_v1", State: "active",
				SourcePermissionID: sourcePermission, SourcePermissionState: "active",
				SourceResourceID: sourceResource, SourceActionID: sourceAction,
				SourceResourceState: "active", SourceActionState: "active",
				ImpliedPermissionID:    testPermissionID,
				ImpliedPermissionState: "active",
				ImpliedResourceID:      testResourceID, ImpliedActionID: testActionID,
				ImpliedResourceState: "active", ImpliedActionState: "active",
			}
			decision := evaluate(testRequest(t),
				testSnapshot(t, []BindingFact{binding}, []ImplicationFact{edge}))
			if decision.Allowed() {
				t.Fatalf("forms record implication granted %s", path)
			}
		})
	}
}

func TestEvaluateRejectsUnknownImplicationKind(t *testing.T) {
	edge := ImplicationFact{ID: uuid.New(), Kind: "unknown_v1", State: "active"}
	if edge.valid() {
		t.Fatal("unknown implication kind is valid")
	}
}
