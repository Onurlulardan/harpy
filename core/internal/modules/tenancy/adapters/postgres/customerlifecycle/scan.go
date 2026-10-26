package customerlifecycle

import (
	"time"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
)

func scanTenant(row Row, classify func(error) error) (lifecycle.Tenant, error) {
	return scanTenantWithPrefix(row, classify)
}

func scanTenantWithPrefix(row Row, classify func(error) error, prefix ...any) (
	lifecycle.Tenant, error,
) {
	var result lifecycle.Tenant
	var offboardedAt, purgedAt *time.Time
	targets := append(prefix, &result.ID, &result.Name, &result.Slug, &result.Status,
		&result.Revision, &result.SettingsRevision, &result.CreatedAt, &result.UpdatedAt,
		&offboardedAt, &purgedAt)
	if err := row.Scan(targets...); err != nil {
		return lifecycle.Tenant{}, classify(err)
	}
	result.CreatedAt, result.UpdatedAt = result.CreatedAt.UTC(), result.UpdatedAt.UTC()
	if offboardedAt != nil {
		result.OffboardedAt = offboardedAt.UTC()
	}
	if purgedAt != nil {
		result.PurgedAt = purgedAt.UTC()
	}
	if result.ID == [16]byte{} || result.Name == "" || result.Slug == "" ||
		result.Revision < 1 || result.SettingsRevision < 1 || result.CreatedAt.IsZero() ||
		result.UpdatedAt.IsZero() {
		return lifecycle.Tenant{}, lifecycle.ErrUnavailable
	}
	return result, nil
}

func scanOperation(row Row, classify func(error) error, prefix ...any) (
	lifecycle.Operation, string, error,
) {
	var result lifecycle.Operation
	var completedAt *time.Time
	var failureCode *string
	targets := make([]any, 0, len(prefix)+11)
	targets = append(targets, prefix...)
	targets = append(targets, &result.ID, &result.ControlTenantID, &result.TargetTenantID,
		&result.Kind, &result.State, &result.Revision, &result.Replay,
		&result.CreatedAt, &result.UpdatedAt, &completedAt, &failureCode)
	if err := row.Scan(targets...); err != nil {
		return lifecycle.Operation{}, "", classify(err)
	}
	result.CreatedAt, result.UpdatedAt = result.CreatedAt.UTC(), result.UpdatedAt.UTC()
	if completedAt != nil {
		result.CompletedAt = completedAt.UTC()
	}
	if failureCode != nil {
		result.FailureCode = *failureCode
	}
	if result.ID == [16]byte{} || result.ControlTenantID == [16]byte{} ||
		result.TargetTenantID == [16]byte{} || result.Kind == "" || result.State == "" ||
		result.Revision < 1 || result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		return lifecycle.Operation{}, "", lifecycle.ErrUnavailable
	}
	return result, "", nil
}
