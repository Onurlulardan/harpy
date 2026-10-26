package purge

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Command struct {
	ids   foundation.IDSource
	clock foundation.Clock
}

type Execution struct {
	Result    Result
	Deletions []Deletion
	More      bool
}

func NewCommand(ids foundation.IDSource, clock foundation.Clock) (*Command, error) {
	if foundation.NilPort(ids) || foundation.NilPort(clock) {
		return nil, fmt.Errorf("documents purge command dependencies are invalid")
	}
	return &Command{ids: ids, clock: clock}, nil
}

func (command *Command) Execute(ctx context.Context, store Store,
	itemID domain.ID,
) (Execution, error) {
	return command.Resume(ctx, store, itemID, domain.ID{},
		domain.MaximumArchiveFiles)
}
