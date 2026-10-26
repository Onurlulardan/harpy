package assignment

import (
	"context"
	"errors"
	"testing"
)

func TestListAssignmentsRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.ListAssignments(context.Background(), TenantScope{}, ListAssignmentsQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid list error = %v", err)
	}
}
