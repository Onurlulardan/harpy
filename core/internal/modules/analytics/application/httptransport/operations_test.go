package httptransport

import "testing"

func TestQueryReadAllowsLatestPublishedOperation(t *testing.T) {
	if !validPermissionOperation(
		PermissionQueryRead,
		OperationQueryGetPublished,
	) {
		t.Fatal("query read must allow the latest published read model")
	}
	if validPermissionOperation(
		PermissionQueryUpdate,
		OperationQueryGetPublished,
	) {
		t.Fatal("query update must not authorize the latest published read model")
	}
}
