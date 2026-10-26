package config

import (
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestWorkerRequiresGoMemoryLimitAtOrAboveMemoryClass(t *testing.T) {
	tests := []struct {
		name  string
		value string
		valid bool
		bytes int64
	}{
		{name: "exact", value: "512MiB", valid: true, bytes: 512 * 1024 * 1024},
		{name: "larger", value: "1GiB", valid: true, bytes: 1024 * 1024 * 1024},
		{name: "exact bytes", value: "536870912B", valid: true, bytes: 512 * 1024 * 1024},
		{name: "below", value: "511MiB"},
		{name: "disabled", value: "off"},
		{name: "fraction", value: "0.5GiB"},
		{name: "whitespace", value: " 512MiB"},
		{name: "overflow", value: "9223372036854775807TiB"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entries := withEnvironment(validWorkerEnvironment(),
				workerMemoryLimitKey+"="+test.value)
			worker, err := loadWorker(entries)
			if test.valid && err != nil {
				t.Fatal(err)
			}
			if !test.valid && (err == nil || !strings.Contains(err.Error(), workerMemoryLimitKey)) {
				t.Fatalf("invalid memory limit result: %v", err)
			}
			if test.valid && worker.MemoryLimit().Bytes() != test.bytes {
				t.Fatalf("memory limit bytes = %d", worker.MemoryLimit().Bytes())
			}
		})
	}
}

func TestWorkerRejectsMissingGoMemoryLimit(t *testing.T) {
	entries := withoutEnvironmentKey(validWorkerEnvironment(), workerMemoryLimitKey)
	if _, err := loadWorker(entries); err == nil ||
		!strings.Contains(err.Error(), workerMemoryLimitKey) {
		t.Fatalf("missing GOMEMLIMIT result: %v", err)
	}
}

func TestWorkerMemoryClassCoversFixedImportReservations(t *testing.T) {
	allocation := runtimev1.FormsImportResourceAllocation()
	if err := allocation.Validate(); err != nil {
		t.Fatal(err)
	}
	if runtimev1.WorkerMinimumMemoryLimitBytes < 2*allocation.TotalBytes() {
		t.Fatalf("memory class %d does not cover fixed import reservation %d",
			runtimev1.WorkerMinimumMemoryLimitBytes, allocation.TotalBytes())
	}
}

func TestWorkerMemoryClassRejectsLowerRuntimeLimit(t *testing.T) {
	worker, err := loadWorker(validWorkerEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.MemoryLimit().ValidateRuntime(
		runtimev1.WorkerMinimumMemoryLimitBytes - 1,
	); err == nil {
		t.Fatal("lower active runtime memory limit was accepted")
	}
	if err := worker.MemoryLimit().ValidateRuntime(
		runtimev1.WorkerMinimumMemoryLimitBytes,
	); err != nil {
		t.Fatal(err)
	}
}
