package runtimev1

import "fmt"

const (
	WorkerMinimumMemoryLimitBytes    int64 = 512 * 1024 * 1024
	FormsImportReservationBytes      int64 = 64 * 1024 * 1024
	FormsImportMaximumConcurrent           = 4
	FormsImportTotalReservationBytes int64 = 256 * 1024 * 1024
)

type WorkerResourceComponent struct {
	name  string
	bytes int64
}

func (component WorkerResourceComponent) Name() string { return component.name }
func (component WorkerResourceComponent) Bytes() int64 { return component.bytes }

type WorkerResourceAllocation struct {
	component         string
	reservationBytes  int64
	maximumConcurrent int
	totalBytes        int64
	components        []WorkerResourceComponent
}

func FormsImportResourceAllocation() WorkerResourceAllocation {
	return WorkerResourceAllocation{
		component:         "forms.record-import",
		reservationBytes:  FormsImportReservationBytes,
		maximumConcurrent: FormsImportMaximumConcurrent,
		totalBytes:        FormsImportTotalReservationBytes,
		components: []WorkerResourceComponent{
			{name: "fcart_decrypt_buffers", bytes: 16 * 1024 * 1024},
			{name: "source_page_and_lookahead", bytes: 9 * 1024 * 1024},
			{name: "mapped_write_page", bytes: 8 * 1024 * 1024},
			{name: "command_event_receipts", bytes: 8 * 1024 * 1024},
			{name: "postgres_driver_buffers", bytes: 8 * 1024 * 1024},
			{name: "parser_canonicalization", bytes: 7 * 1024 * 1024},
			{name: "allocator_headroom", bytes: 8 * 1024 * 1024},
		},
	}
}

func (allocation WorkerResourceAllocation) Validate() error {
	if allocation.component != "forms.record-import" ||
		allocation.reservationBytes != FormsImportReservationBytes ||
		allocation.maximumConcurrent != FormsImportMaximumConcurrent ||
		allocation.totalBytes != FormsImportTotalReservationBytes ||
		allocation.reservationBytes*int64(allocation.maximumConcurrent) != allocation.totalBytes {
		return fmt.Errorf("worker resource allocation is invalid")
	}
	var total int64
	for _, component := range allocation.components {
		if component.name == "" || component.bytes < 1 {
			return fmt.Errorf("worker resource component is invalid")
		}
		total += component.bytes
	}
	if total != allocation.reservationBytes {
		return fmt.Errorf("worker resource components do not match reservation")
	}
	return nil
}

func (allocation WorkerResourceAllocation) Component() string { return allocation.component }
func (allocation WorkerResourceAllocation) ReservationBytes() int64 {
	return allocation.reservationBytes
}
func (allocation WorkerResourceAllocation) MaximumConcurrent() int {
	return allocation.maximumConcurrent
}
func (allocation WorkerResourceAllocation) TotalBytes() int64 { return allocation.totalBytes }
func (allocation WorkerResourceAllocation) Components() []WorkerResourceComponent {
	return append([]WorkerResourceComponent(nil), allocation.components...)
}
