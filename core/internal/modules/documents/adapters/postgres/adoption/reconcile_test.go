package adoption

import "testing"

func TestReconcileOperationContract(t *testing.T) {
	if any((*Store).Reconcile) == nil {
		t.Fatal("adoption reconcile persistence is unavailable")
	}
}
