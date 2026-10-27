package adoption

import (
	"strings"
	"testing"
)

func TestCompilerFencesUnsafeV1Shapes(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Snapshot, *Mapping)
		code   string
	}{
		{"organization cycle", func(s *Snapshot, _ *Mapping) {
			s.Organizations[0].ParentID = s.Organizations[0].ID.String()
		}, "organization_cycle"},
		{"malformed role", func(s *Snapshot, _ *Mapping) {
			s.Users[0].RoleIDs = []string{"not-a-uuid"}
		}, "malformed_role_id"},
		{"invalid bcrypt", func(s *Snapshot, _ *Mapping) {
			s.Users[0].PasswordHash = "$2a$04$short"
		}, "credential_format_invalid"},
		{"direct wildcard", func(s *Snapshot, _ *Mapping) {
			s.Permissions[0].Target, s.Permissions[0].RoleID = "USER", nil
			s.Permissions[0].UserID = pointer(s.Users[0].ID)
		}, "wildcard_review_required"},
		{"missing permission map", func(_ *Snapshot, m *Mapping) {
			m.Permissions = nil
		}, "permission_mapping_missing"},
		{"missing profile map", func(_ *Snapshot, m *Mapping) {
			m.ProfileFields = nil
		}, "profile_field_mapping_missing"},
		{"missing layout map", func(_ *Snapshot, m *Mapping) {
			m.LayoutFields = nil
		}, "layout_field_mapping_missing"},
		{"missing invitation reissue", func(_ *Snapshot, m *Mapping) {
			m.Invitations = nil
		}, "invitation_reissue_mapping_missing"},
	}
	compiler, err := NewCompiler(testNormalizer, testNormalizer)
	if err != nil {
		t.Fatalf("create compiler: %v", err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot, mapping := validFixture()
			test.mutate(&snapshot, &mapping)
			plan, compileErr := compiler.Evaluate(snapshot, mapping)
			if compileErr != nil {
				t.Fatalf("compile fenced fixture: %v", compileErr)
			}
			if plan.Eligible || !containsIssue(plan.Issues, test.code) {
				t.Fatalf("missing %s fence: %+v", test.code, plan.Issues)
			}
		})
	}
}

func containsIssue(values []Issue, code string) bool {
	for _, value := range values {
		if strings.EqualFold(value.Code, code) {
			return true
		}
	}
	return false
}
