package httptransport

import "testing"

func TestDashboardExportOperationsRequireDedicatedPermission(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{OperationDashboardExportCreate,
		OperationDashboardExportGet, OperationDashboardExportDelete,
		OperationDashboardExportContent} {
		if !validPermissionOperation(PermissionDashboardExport, operation) {
			t.Fatalf("dashboard export operation %s was not authorized by its permission", operation)
		}
		if validPermissionOperation(PermissionDashboardRead, operation) ||
			validPermissionOperation(PermissionDashboardUpdate, operation) {
			t.Fatalf("dashboard export operation %s accepted a broader permission", operation)
		}
	}
	if validPermissionOperation(PermissionDashboardExport, OperationDashboardGet) {
		t.Fatal("dashboard export permission authorized an unrelated dashboard operation")
	}
}
