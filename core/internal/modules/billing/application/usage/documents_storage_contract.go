package usage

import "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"

func DocumentsStorageEntitlement() (entitlement.Key, error) {
	key, _ := entitlement.NewKey("documents.storage_bytes")
	entry, found := entitlement.RegistryV1ByKey()[key]
	definition := entry.Definition
	if !found || entry.Owner != "documents" || entry.Unit != entitlement.UnitByte ||
		definition.ValueKind() != entitlement.Integer ||
		definition.MeterKind() != entitlement.Gauge ||
		definition.ResetPolicy() != entitlement.Never ||
		definition.Enforcement() != entitlement.Hard ||
		!definition.AllowsUnlimited() {
		return "", ErrConfiguration
	}
	return key, nil
}
