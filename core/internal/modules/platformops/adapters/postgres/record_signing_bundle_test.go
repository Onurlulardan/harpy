package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestRecordSigningBundle(t *testing.T) {
	database := &operationDatabase{}
	err := operationStore(database).RecordSigningBundle(context.Background(), ports.SigningBundleRecord{
		EnvironmentID: "018f0000-0000-7000-8000-000000000001",
		Generation:    1, SHA256: strings.Repeat("a", 64),
		BundleBytes: []byte("bundle"), CreatedAt: time.Now().UTC(),
	})
	if err != nil || database.executions != 1 {
		t.Fatalf("save signing bundle executions=%d err=%v", database.executions, err)
	}
}
