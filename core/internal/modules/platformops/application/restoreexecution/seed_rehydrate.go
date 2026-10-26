package restoreexecution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) SeedRehydrate(
	ctx context.Context,
	record ports.RehydrateControlSeed,
) error {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil ||
		!matchesDigest(record.TrustSHA256, record.TrustBytes) ||
		!matchesDigest(record.PlanSHA256, record.PlanBytes) ||
		!matchesDigest(record.ManifestSHA256, record.ManifestBytes) {
		return ErrTransition
	}
	value, err := service.store.SeedRehydrateControl(ctx, record)
	return changed(err, value)
}

func matchesDigest(expected string, document []byte) bool {
	digest := sha256.Sum256(document)
	return digestPattern.MatchString(expected) &&
		expected == hex.EncodeToString(digest[:])
}
