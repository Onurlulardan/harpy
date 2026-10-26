package postgres

import (
	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func (stamp authorityStamp) matchesConsumerScope(tenantID *string) bool {
	return tenantID != nil && *tenantID == stamp.scopeID.String()
}

func (stamp authorityStamp) matchesConsumerAuthority(value eventsv1.Authority) bool {
	return value.ScopeKind == eventsv1.ScopeTenant && value.ScopeID == stamp.scopeID.String() &&
		value.GroupID == stamp.groupID && value.GroupGeneration == stamp.groupGeneration &&
		value.Domain == stamp.domain && value.CutoverGeneration == stamp.generation &&
		value.WriterOrigin == eventsv1.WriterV2
}

func (stamp authorityStamp) matchesEffectAuthority(value effectsv1.Authority) bool {
	return value.ScopeKind == effectsv1.ScopeTenant && value.ScopeID == stamp.scopeID.String() &&
		value.GroupID == stamp.groupID && value.GroupGeneration == stamp.groupGeneration &&
		value.Domain == stamp.domain && value.CutoverGeneration == stamp.generation &&
		value.WriterOrigin == effectsv1.WriterV2
}

func (stamp authorityStamp) eventAuthority() eventsv1.Authority {
	return eventsv1.Authority{
		ScopeKind: eventsv1.ScopeTenant, ScopeID: stamp.scopeID.String(),
		GroupID: stamp.groupID, GroupGeneration: stamp.groupGeneration,
		Domain: stamp.domain, CutoverGeneration: stamp.generation,
		WriterOrigin: eventsv1.WriterV2,
	}
}

func consumerAuthorityArguments(stamp authorityStamp) []any {
	return []any{
		"tenant", stamp.scopeID, stamp.groupID,
		stamp.groupGeneration, stamp.domain, stamp.generation,
	}
}
