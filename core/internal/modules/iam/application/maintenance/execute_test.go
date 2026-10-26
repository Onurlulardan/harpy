package maintenance

import (
	"context"
	"errors"
	"testing"
)

func TestExecuteRejectsNilCommandDispatcher(t *testing.T) {
	var commands *Commands
	if _, err := commands.Execute(context.Background(), Claim{}); !errors.Is(
		err, ErrCommandInvalid,
	) {
		t.Fatalf("nil dispatcher error = %v", err)
	}
}
