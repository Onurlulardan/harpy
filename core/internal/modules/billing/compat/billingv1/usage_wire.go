package billingv1

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

type usageCounterWire struct {
	Current int64 `json:"current"`
	Limit   int64 `json:"limit"`
}

type usageStorageWire struct {
	CurrentMB int64 `json:"current_mb"`
	LimitMB   int64 `json:"limit_mb"`
}

type usageWire struct {
	Submissions usageCounterWire `json:"submissions"`
	Users       usageCounterWire `json:"users"`
	Storage     usageStorageWire `json:"storage"`
	Limits      planLimits       `json:"limits"`
}

func legacyUsage(view billingapi.UsageView) usageWire {
	result := usageWire{}
	for _, meter := range view.Meters {
		limit := int64(-1)
		if !meter.Unlimited && meter.Limit != nil {
			limit = *meter.Limit
		}
		switch meter.EntitlementKey {
		case "iam.active_users":
			result.Users = usageCounterWire{meter.Current, limit}
			result.Limits.MaxUsers = limit
		case "documents.storage_bytes":
			result.Storage.CurrentMB = meter.Current / (1024 * 1024)
			if limit == -1 {
				result.Storage.LimitMB, result.Limits.MaxStorageMB = -1, -1
			} else {
				result.Storage.LimitMB = limit / (1024 * 1024)
				result.Limits.MaxStorageMB = result.Storage.LimitMB
			}
		}
	}
	return result
}
