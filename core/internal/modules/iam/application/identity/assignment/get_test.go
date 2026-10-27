package assignment

import (
	"context"
	"errors"
	"testing"
)

func TestGetAssignmentRejectsInvalidQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.GetAssignment(context.Background(), TenantScope{}, GetAssignmentQuery{}); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("invalid get error = %v", err)
	}
}
