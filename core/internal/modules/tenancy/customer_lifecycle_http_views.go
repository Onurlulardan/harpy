package tenancy

import (
	"time"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
)

type customerLifecycleTenantDocument struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Slug             string     `json:"slug"`
	Status           string     `json:"status"`
	Revision         int64      `json:"lifecycle_revision"`
	SettingsRevision int64      `json:"settings_revision"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	OffboardedAt     *time.Time `json:"offboarded_at,omitempty"`
	PurgedAt         *time.Time `json:"purged_at,omitempty"`
}

func customerLifecycleTenantView(tenant lifecycle.Tenant) customerLifecycleTenantDocument {
	result := customerLifecycleTenantDocument{ID: tenant.ID.String(), Name: tenant.Name,
		Slug: tenant.Slug, Status: tenant.Status, Revision: tenant.Revision,
		SettingsRevision: tenant.SettingsRevision,
		CreatedAt:        tenant.CreatedAt, UpdatedAt: tenant.UpdatedAt}
	if !tenant.OffboardedAt.IsZero() {
		value := tenant.OffboardedAt
		result.OffboardedAt = &value
	}
	if !tenant.PurgedAt.IsZero() {
		value := tenant.PurgedAt
		result.PurgedAt = &value
	}
	return result
}

type customerLifecyclePageDocument struct {
	Items      []customerLifecycleTenantDocument `json:"items"`
	NextCursor string                            `json:"next_cursor,omitempty"`
}

func customerLifecyclePageView(page lifecycle.Page) customerLifecyclePageDocument {
	result := customerLifecyclePageDocument{
		Items: make([]customerLifecycleTenantDocument, 0, len(page.Items)),
	}
	for _, tenant := range page.Items {
		result.Items = append(result.Items, customerLifecycleTenantView(tenant))
	}
	if page.NextCursor != [16]byte{} {
		result.NextCursor = page.NextCursor.String()
	}
	return result
}

type customerLifecycleOperationDocument struct {
	OperationID    string     `json:"operation_id"`
	TenantID       string     `json:"tenant_id"`
	Kind           string     `json:"kind"`
	State          string     `json:"state"`
	ResultRevision int64      `json:"result_revision"`
	Replay         bool       `json:"replay"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	FailureCode    string     `json:"failure_code,omitempty"`
	Location       string     `json:"location,omitempty"`
}

func customerLifecycleOperationView(operation lifecycle.Operation,
	location string,
) customerLifecycleOperationDocument {
	result := customerLifecycleOperationDocument{OperationID: operation.ID.String(),
		TenantID: operation.TargetTenantID.String(), Kind: operation.Kind,
		State: operation.State, ResultRevision: operation.Revision, Replay: operation.Replay,
		CreatedAt: operation.CreatedAt, UpdatedAt: operation.UpdatedAt,
		FailureCode: operation.FailureCode, Location: location}
	if !operation.CompletedAt.IsZero() {
		value := operation.CompletedAt
		result.CompletedAt = &value
	}
	return result
}

type customerLifecycleLegacyPageDocument struct {
	Data  []customerLifecycleTenantDocument `json:"data"`
	Total int                               `json:"total"`
	Next  string                            `json:"next_cursor,omitempty"`
}

func customerLifecycleLegacyPageView(page lifecycle.Page) customerLifecycleLegacyPageDocument {
	canonical := customerLifecyclePageView(page)
	return customerLifecycleLegacyPageDocument{Data: canonical.Items,
		Total: len(canonical.Items), Next: canonical.NextCursor}
}

type customerLifecycleLegacyTenantView struct {
	Data customerLifecycleTenantDocument `json:"data"`
}

type customerLifecycleLegacyMutationView struct {
	Message string                          `json:"message"`
	Data    customerLifecycleTenantDocument `json:"data"`
}
