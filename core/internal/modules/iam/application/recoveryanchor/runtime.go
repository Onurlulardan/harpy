package recoveryanchor

import (
	"context"
	"fmt"
	"time"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const recoveryAnchorChallengeLifetime = 5 * time.Minute

type RuntimePlatform interface {
	ResolveSnapshotDomainGenerations(
		context.Context, uuid.UUID, []string,
	) (map[string]int64, error)
}

type Runtime struct {
	pool        RuntimePlatform
	coordinator *platformtx.Coordinator
	service     *Service
	verifier    anchorapi.RegistrationVerifier
	rpID        string
	origins     []string
	now         func() time.Time
}

func NewRuntime(
	pool RuntimePlatform,
	coordinator *platformtx.Coordinator,
	service *Service,
	verifier anchorapi.RegistrationVerifier,
	rpID string,
	origins []string,
) (*Runtime, error) {
	if pool == nil || coordinator == nil || service == nil || verifier == nil || rpID == "" ||
		len(origins) == 0 {
		return nil, fmt.Errorf("recovery anchor runtime dependencies are incomplete")
	}
	return &Runtime{pool: pool, coordinator: coordinator, service: service,
		verifier: verifier,
		rpID:     rpID, origins: append([]string(nil), origins...),
		now: func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }}, nil
}
