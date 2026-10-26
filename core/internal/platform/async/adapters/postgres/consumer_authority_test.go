package postgres

import (
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func TestConsumerAuthorityIsExactAndTenantBound(t *testing.T) {
	t.Parallel()
	stamp := authorityStamp{
		scopeID: authorityFixture().scopeID, groupID: "operations",
		groupGeneration: 1, domain: "workflow", generation: 1,
	}
	if !stamp.matchesConsumerScope(stringPointerValue(consumerTenantA)) ||
		!stamp.matchesConsumerAuthority(consumerEventAuthority(consumerTenantA)) ||
		!stamp.matchesEffectAuthority(consumerEffectAuthority(consumerTenantA)) {
		t.Fatal("exact consumer authority was rejected")
	}
	wrongTenant := consumerTenantB
	wrongEvent := consumerEventAuthority(consumerTenantA)
	wrongEvent.GroupGeneration++
	wrongEffect := consumerEffectAuthority(consumerTenantA)
	wrongEffect.Domain = "notifications"
	if stamp.matchesConsumerScope(&wrongTenant) || stamp.matchesConsumerAuthority(wrongEvent) ||
		stamp.matchesEffectAuthority(wrongEffect) {
		t.Fatal("cross-tenant or stale consumer authority was accepted")
	}
}

func TestConsumerAuthorityRejectsWireOriginDowngrade(t *testing.T) {
	t.Parallel()
	stamp := consumerRepositoryFixture(eventsv1.ProducerCatalog{}, eventsv1.ConsumerCatalog{}).authority
	eventAuthority := consumerEventAuthority(consumerTenantA)
	eventAuthority.WriterOrigin = eventsv1.WriterV1
	effectAuthority := consumerEffectAuthority(consumerTenantA)
	effectAuthority.WriterOrigin = effectsv1.WriterV1
	if stamp.matchesConsumerAuthority(eventAuthority) || stamp.matchesEffectAuthority(effectAuthority) {
		t.Fatal("V1 consumer authority was accepted")
	}
}

func stringPointerValue(value string) *string { return &value }
