package restoreexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestCompleteBindsReceiptBytes(t *testing.T) {
	document := []byte("signed receipt")
	digest := sha256.Sum256(document)
	record := ports.RestoreCompletion{RestoreExecutionIdentity: fixtureIdentity,
		ReceiptSHA256: hex.EncodeToString(digest[:]), ReceiptBytes: document}
	if err := New(&storeFixture{changed: true}).Complete(t.Context(), record); err != nil {
		t.Fatal(err)
	}
}
