package recoveryevidence

import (
	"context"

	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports/recoveryevidence"
)

type Store interface {
	ReadRecoveryHead(context.Context, string) (evidenceports.Head, error)
}
