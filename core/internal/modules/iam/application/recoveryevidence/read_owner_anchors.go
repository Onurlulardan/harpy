package recoveryevidence

import (
	"context"
	"fmt"

	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/recoveryevidence"
)

func (service *Service) ReadOwnerAnchors(
	ctx context.Context,
	environmentID string,
) ([]evidenceports.Anchor, error) {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || environmentID == "" {
		return nil, fmt.Errorf("IAM recovery evidence request is invalid")
	}
	return service.store.ReadOwnerAnchors(ctx, environmentID)
}

var _ evidenceports.Reader = (*Service)(nil)
