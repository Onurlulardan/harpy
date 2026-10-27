package iam

import "testing"

func TestFormsLifecycleAccessMapsOnlyExactPermissionOperations(t *testing.T) {
	tests := [][4]string{
		{"forms.definition.read", "forms.definition.get", "forms.definition", "read"},
		{"forms.definition.update", "forms.definition.create_version", "forms.definition", "update"},
		{"forms.definition.retire", "forms.definition.retire", "forms.definition", "retire"},
		{"forms.publication.plan", "forms.publication.plan", "forms.publication", "plan"},
		{"forms.publication.deactivate", "forms.publication.deactivate", "forms.publication", "deactivate"},
	}
	for _, test := range tests {
		target, found := resolveFormsLifecycleAccess(test[0], test[1])
		if !found || target.resource != test[2] || target.action != test[3] {
			t.Fatalf("target %s/%s = %#v", test[0], test[1], target)
		}
	}
	for _, rejected := range [][2]string{
		{"forms.definition.read", "forms.definition.create"},
		{"forms.publication.plan", "forms.publication.activate"},
		{"forms.publication.read", "forms.publication.plan"},
	} {
		if _, found := resolveFormsLifecycleAccess(rejected[0], rejected[1]); found {
			t.Fatalf("unexpected lifecycle target: %v", rejected)
		}
	}
}
