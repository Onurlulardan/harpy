package membership

import (
	"context"
	"errors"
	"testing"
)

func TestListMembershipsRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.ListMemberships(context.Background(), TenantScope{}, ListMembershipsQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid list error = %v", err)
	}
}
