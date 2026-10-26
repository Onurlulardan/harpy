package compatibility

import (
	"context"
	"time"
)

const legacySilence = 30 * 24 * time.Hour

func (application *Application) RetireEdge(
	ctx context.Context, mapping LegacyMapping,
) (LegacyMapping, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		application.lifecycle == nil || !validMapping(mapping) ||
		mapping.State != "active" || mapping.ConsumersAckedAt == nil {
		return LegacyMapping{}, ErrInvalid
	}
	now := application.clock.Now()
	evidence, err := application.authority.Evidence(ctx)
	if err != nil || !validTime(now) || !validRetirementEvidence(evidence, now) ||
		now.Before(mapping.RemoveAfter) || (mapping.LastHitAt != nil &&
		now.Sub(*mapping.LastHitAt) < legacySilence) {
		return LegacyMapping{}, ErrBlocked
	}
	result, err := application.lifecycle.Retire(ctx, mapping, evidence, now)
	if err != nil {
		return LegacyMapping{}, err
	}
	if !sameMappingIdentity(result, mapping) || result.State != "retired" {
		return LegacyMapping{}, ErrUnavailable
	}
	return result, nil
}

func validRetirementEvidence(value RetirementEvidence, now time.Time) bool {
	return value.AllTenantsV2 && value.SHA256 != [32]byte{} && validTime(value.ObservedAt) &&
		!value.ObservedAt.After(now) && now.Sub(value.ObservedAt) <= 5*time.Minute
}
