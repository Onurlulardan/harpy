package sessionview

import (
	"context"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestReadRejectsClosedScope(t *testing.T) {
	port, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := port.Read(
		context.Background(), platformtx.TenantScope{}, uuid.UUID{1}, uuid.UUID{2},
	); err == nil {
		t.Fatal("closed scope was accepted")
	}
}
