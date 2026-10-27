package lifecyclehttp

import "testing"

func TestReadOperation(t *testing.T) {
	if err := exerciseOperation("get_operation", "catalog.operation.get",
		func(handler *Handler) Route { return handler.ReadOperationRoute() }); err != nil {
		t.Fatal(err)
	}
}
