package runtimev1

import "testing"

func TestFormsImportResourceAllocationIsExactAndSelfConsistent(t *testing.T) {
	allocation := FormsImportResourceAllocation()
	if err := allocation.Validate(); err != nil {
		t.Fatal(err)
	}
	if WorkerMinimumMemoryLimitBytes != 512*1024*1024 ||
		allocation.Component() != "forms.record-import" ||
		allocation.ReservationBytes() != 64*1024*1024 ||
		allocation.MaximumConcurrent() != 4 ||
		allocation.TotalBytes() != 256*1024*1024 {
		t.Fatalf("unexpected worker resource allocation: %#v", allocation)
	}
	seen := make(map[string]struct{})
	var componentBytes int64
	for _, component := range allocation.Components() {
		if _, duplicate := seen[component.Name()]; duplicate {
			t.Fatalf("duplicate allocation component %q", component.Name())
		}
		seen[component.Name()] = struct{}{}
		componentBytes += component.Bytes()
	}
	if componentBytes != allocation.ReservationBytes() {
		t.Fatalf("component bytes = %d", componentBytes)
	}
}
