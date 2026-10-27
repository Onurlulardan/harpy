package systemtransfer

import (
	"errors"
	"testing"
)

func TestExecuteExportRejectsMissingWorker(t *testing.T) {
	var worker *ExportWorker
	result, err := worker.ExecuteExport(t.Context(), ExportJobClaim{})
	if !errors.Is(err, ErrExportWorkerInvalid) || result != (ExportExecutionResult{}) {
		t.Fatalf("ExecuteExport() = %#v, %v", result, err)
	}
}
