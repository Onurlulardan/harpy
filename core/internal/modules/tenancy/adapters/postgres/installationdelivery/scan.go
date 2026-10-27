package installationdelivery

import (
	"time"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func scanDelivery(row Row) (installationdelivery.Delivery, error) {
	var result installationdelivery.Delivery
	var state string
	var adapter *string
	var lease, acknowledged, deleted, completed *time.Time
	err := row.Scan(&result.OperationID, &result.TenantID, &result.FirstOwnerUserID,
		&result.CurrentAttemptID, &result.CurrentCredentialID,
		&result.CurrentCredentialVersion, &state, &result.Revision,
		&result.DatabaseNow, &adapter, &lease, &acknowledged, &deleted, &completed)
	if err != nil {
		return installationdelivery.Delivery{}, installationdelivery.ErrUnavailable
	}
	result.DatabaseNow = result.DatabaseNow.UTC()
	result.State = installationdelivery.State(state)
	if adapter != nil {
		result.ClaimAdapterID = *adapter
	}
	result.ClaimLeaseUntil = nullableTime(lease)
	result.AcknowledgedAt = nullableTime(acknowledged)
	result.SecretDeletedAt = nullableTime(deleted)
	result.CompletedAt = nullableTime(completed)
	return result, nil
}

func nullableTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	result := value.UTC()
	return &result
}
