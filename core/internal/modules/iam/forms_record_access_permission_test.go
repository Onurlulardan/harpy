package iam

import "testing"

func TestFormsRecordAccessMapsOnlyExactActionOperations(t *testing.T) {
	tests := []struct{ action, operation, want string }{
		{"create", "forms.record_kernel.create", "create"},
		{"read", "forms.record_kernel.get", "read"},
		{"read", "restquery.execute", "read"},
		{"update", "forms.record_kernel.update", "update"},
		{"delete", "forms.record_kernel.delete", "delete"},
		{"export", "forms.record_export.submit", "export"},
		{"export", "forms.record_export.get", "export"},
		{"export", "forms.record_export.delete", "export"},
		{"export", "forms.record_export.download", "export"},
		{"import", "forms.record_import.create_source", "import"},
		{"import", "forms.record_import.load_source", "import"},
		{"import", "forms.record_import.submit", "import"},
		{"import", "forms.record_import.get", "import"},
		{"import", "forms.record_import.list", "import"},
		{"import", "forms.record_import.cancel", "import"},
		{"import", "forms.record_import.resume", "import"},
		{"import", "forms.record_import.list_errors", "import"},
	}
	for _, test := range tests {
		target, found := resolveFormsRecordAccess(test.action, test.operation)
		if !found || target.action != test.want {
			t.Fatalf("target %s/%s = %#v", test.action, test.operation, target)
		}
	}
	for _, rejected := range [][2]string{
		{"read", "forms.record_kernel.create"},
		{"*", "restquery.execute"},
		{"read", "forms.record_kernel.list"},
		{"read", ""},
	} {
		if _, found := resolveFormsRecordAccess(rejected[0], rejected[1]); found {
			t.Fatalf("unexpected Forms record access target: %v", rejected)
		}
	}
}
