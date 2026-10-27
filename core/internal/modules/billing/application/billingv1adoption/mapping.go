package billingv1adoptionapp

import (
	"time"

	"github.com/google/uuid"
)

type GlobalMapping struct {
	BasePlanID                 uuid.UUID
	GlobalCouponEvidenceSHA256 map[uuid.UUID]string
}

type TenantMapping struct {
	TenantID                 uuid.UUID
	PastDueGraceEndsAt       map[uuid.UUID]time.Time
	UsageFlushEvidenceSHA256 map[uuid.UUID]string
}

type AdoptionMapping struct {
	ContractVersion string
	SourceTimezone  string
	Global          GlobalMapping
	Tenants         []TenantMapping
}

const MappingContractVersion = "harpy/billing-v1-adoption-mapping/v1"

func (mapping AdoptionMapping) tenant(tenantID uuid.UUID) (TenantMapping, bool) {
	for _, value := range mapping.Tenants {
		if value.TenantID == tenantID {
			return value, true
		}
	}
	return TenantMapping{}, false
}
