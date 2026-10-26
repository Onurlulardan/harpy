package billing

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const SaaSFreeBaseOfferingKey = "saas_free_base_v1"

var ErrSaaSFreeBaseUnavailable = errors.New("billing SaaS free base is unavailable")

var saaSFreeBaseSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)

type SaaSFreeBaseReader interface {
	FetchSaaSFreeBase(context.Context) (SaaSFreeBasePin, error)
}

type SaaSFreeBasePin struct {
	VersionID   uuid.UUID
	Revision    int64
	SHA256      string
	EffectiveAt time.Time
}

func (pin SaaSFreeBasePin) IsValid() bool {
	return pin.VersionID != uuid.Nil && pin.VersionID.Version() == 7 &&
		pin.Revision == 1 && saaSFreeBaseSHA256.MatchString(pin.SHA256) &&
		!pin.EffectiveAt.IsZero() && pin.EffectiveAt.Location() == time.UTC
}

func ReadSaaSFreeBase(
	ctx context.Context,
	reader SaaSFreeBaseReader,
) (SaaSFreeBasePin, error) {
	if ctx == nil || ctx.Err() != nil || reader == nil {
		return SaaSFreeBasePin{}, ErrSaaSFreeBaseUnavailable
	}
	pin, err := reader.FetchSaaSFreeBase(ctx)
	if err != nil || !pin.IsValid() {
		return SaaSFreeBasePin{}, fmt.Errorf("read SaaS free base: %w",
			errors.Join(ErrSaaSFreeBaseUnavailable, err))
	}
	return pin, nil
}
