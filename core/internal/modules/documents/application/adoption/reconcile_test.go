package adoption

import "testing"

func TestReconcileOperationContract(t *testing.T) {
	if any((*Application).Reconcile) == nil {
		t.Fatal("reconcile operation is unavailable")
	}
}
