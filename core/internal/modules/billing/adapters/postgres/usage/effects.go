package usage

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

type Effects interface {
	PublishThreshold(context.Context, usageapp.ThresholdEventIntent) error
}
