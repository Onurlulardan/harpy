package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type EffectStore interface {
	LockEffect(context.Context, string, [32]byte, domain.ID) error
	FindEffect(context.Context, string, [32]byte, domain.ID) (domain.EffectReceipt, bool, error)
	InsertEffect(context.Context, domain.EffectReceipt) error
}
