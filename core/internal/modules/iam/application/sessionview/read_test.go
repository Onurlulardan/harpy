package sessionview

import (
	"context"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestReadFailsClosedOutsideActorTenantScope(t *testing.T) {
	reader, err := NewReader(rejectedPort{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reader.Read(
		context.Background(), platformtx.TenantScope{}, uuid.UUID{1}, uuid.UUID{2},
	); err == nil {
		t.Fatal("session view escaped the actor tenant scope")
	}
}

type rejectedPort struct{}

func (rejectedPort) Read(
	context.Context, platformtx.TenantScope, uuid.UUID, uuid.UUID,
) (View, error) {
	return View{}, ErrStorage
}
