package purge

import "testing"

func TestExecuteOperationContract(t *testing.T) {
	if any((*Command).Execute) == nil {
		t.Fatal("purge execute operation is unavailable")
	}
}
