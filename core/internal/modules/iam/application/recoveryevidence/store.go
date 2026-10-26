package recoveryevidence

import (
	"context"

	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/recoveryevidence"
)

type Store interface {
	ReadOwnerAnchors(context.Context, string) ([]evidenceports.Anchor, error)
}
