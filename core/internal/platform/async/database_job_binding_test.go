package async

import (
	"context"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type bindingCommand struct{}

func TestDatabaseJobBindingTransactionMode(t *testing.T) {
	build := func(platformtx.JobBoundExecutor) DatabaseJobCommand {
		return bindingCommand{}
	}
	regular, err := NewDatabaseJobBinding("documents", build)
	if err != nil || regular.RepeatableRead() {
		t.Fatalf("regular binding = %+v/%v", regular, err)
	}
	repeatable, err := NewRepeatableReadDatabaseJobBinding("documents", build)
	if err != nil || !repeatable.RepeatableRead() || repeatable.Owner() != "documents" {
		t.Fatalf("repeatable binding = %+v/%v", repeatable, err)
	}
}

func (bindingCommand) Execute(
	context.Context, DatabaseJobClaim,
) (DatabaseJobOutcome, error) {
	return CompleteDatabaseJob(), nil
}
