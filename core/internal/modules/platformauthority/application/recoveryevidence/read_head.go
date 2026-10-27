package recoveryevidence

import (
	"context"
	"fmt"

	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports/recoveryevidence"
)

func (service *Service) ReadRecoveryHead(
	ctx context.Context,
	environmentID string,
) (evidenceports.Head, error) {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || environmentID == "" {
		return evidenceports.Head{},
			fmt.Errorf("platform authority recovery evidence request is invalid")
	}
	return service.store.ReadRecoveryHead(ctx, environmentID)
}

var _ evidenceports.Reader = (*Service)(nil)
