package lifecyclehttp

import "testing"

func TestRetryOperation(t *testing.T) {
	if err := exerciseOperation("retry_operation", "catalog.operation.retry",
		func(handler *Handler) Route { return handler.RetryOperationRoute() }); err != nil {
		t.Fatal(err)
	}
}
