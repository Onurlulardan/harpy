package operator

import (
	"encoding/json"
	"io"
	"time"
)

type postRestoreObserved struct {
	Schema             string `json:"schema"`
	EnvironmentID      string `json:"environment_id"`
	RestoreOperationID string `json:"restore_operation_id"`
	State              string `json:"state"`
	ReceiptSHA256      string `json:"receipt_sha256"`
	ReactivatedAt      string `json:"reactivated_at"`
}

func writePostRestoreObserved(
	writer io.Writer,
	environmentID string,
	restoreOperationID string,
	receiptSHA256 string,
	reactivatedAt time.Time,
) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(postRestoreObserved{
		Schema:        "operator-post-restore-reactivation-observed.v1",
		EnvironmentID: environmentID, RestoreOperationID: restoreOperationID,
		State: "reactivated", ReceiptSHA256: receiptSHA256,
		ReactivatedAt: reactivatedAt.UTC().Format(time.RFC3339Nano),
	})
}
