package iam

import "testing"

func TestSystemTransferAccessTargetIsClosedToExactPermissionOperation(t *testing.T) {
	for operation, action := range map[string]string{
		SystemExportSubmitOperation: "export", SystemExportDownloadOperation: "export",
		SystemImportSourceOperation: "import", SystemImportReportOperation: "import",
	} {
		permission := SystemTransferUsers.Permission(action)
		target, found := resolveSystemTransferAccess(
			SystemTransferUsers, permission, operation)
		if !found || target.action != action {
			t.Fatalf("operation=%s target=%+v found=%t", operation, target, found)
		}
	}
	for _, test := range []struct {
		permission, operation string
	}{
		{"iam.users.read", SystemExportSubmitOperation},
		{"iam.roles.export", SystemExportSubmitOperation},
		{"iam.users.export", "iam.record_export.unknown"},
	} {
		if _, found := resolveSystemTransferAccess(
			SystemTransferUsers, test.permission, test.operation); found {
			t.Fatalf("accepted permission=%s operation=%s", test.permission, test.operation)
		}
	}
}
