package migratoroutput

import (
	"io"
	"testing"

	platformmigrator "github.com/Onurlulardan/harpy/core/internal/platform/migrator"
)

type shortWriter struct{}

func (shortWriter) Write(document []byte) (int, error) {
	return len(document) - 1, nil
}

func TestWriteRejectsPartialOutput(t *testing.T) {
	result, err := FromUp(platformmigrator.UpResult{
		RunID:           "11111111-1111-4111-8111-111111111111",
		AppliedVersions: []int64{}, DatabaseVersion: 1,
	})
	if err != nil {
		t.Fatalf("map up: %v", err)
	}
	if err := result.Write(shortWriter{}); err == nil {
		t.Fatal("partial writer accepted")
	}
	if err := result.Write(io.Discard); err != nil {
		t.Fatalf("full writer rejected: %v", err)
	}
}
