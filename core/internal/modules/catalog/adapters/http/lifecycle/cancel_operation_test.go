package lifecyclehttp

import "testing"

func TestCancelOperation(t *testing.T) {
	if err := exerciseOperation("cancel_operation", "catalog.operation.cancel",
		func(handler *Handler) Route { return handler.CancelOperationRoute() }); err != nil {
		t.Fatal(err)
	}
}
