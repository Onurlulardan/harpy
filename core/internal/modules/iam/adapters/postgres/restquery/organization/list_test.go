package organizationpostgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListUsesBoundProfileReaderAndRejectsClosedScope(t *testing.T) {
	port, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := port.List(context.Background(), transaction.TenantScope{}); err == nil {
		t.Fatal("closed scope was accepted")
	}
}
