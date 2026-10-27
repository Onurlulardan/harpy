package customerlifecycle

import (
	"context"
	"encoding/json"
	"time"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/google/uuid"
)

type tenantDocument struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	Slug             string     `json:"slug"`
	Status           string     `json:"status"`
	Revision         int64      `json:"revision"`
	SettingsRevision int64      `json:"settings_revision"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	OffboardedAt     *time.Time `json:"offboarded_at"`
	PurgedAt         *time.Time `json:"purged_at"`
}

func (store *Store) List(ctx context.Context, request lifecycle.ListStoreRequest) (lifecycle.Page, error) {
	if !store.usable(ctx) || request.EnvironmentID == uuid.Nil || request.Limit < 1 || request.Limit > 100 {
		return lifecycle.Page{}, lifecycle.ErrInvalid
	}
	var document []byte
	var next *uuid.UUID
	if err := store.queryRow(ctx, listSQL, request.EnvironmentID, request.Limit,
		nullUUID(request.After)).Scan(&document, &next); err != nil {
		return lifecycle.Page{}, store.classify(err)
	}
	var values []tenantDocument
	if len(document) == 0 || json.Unmarshal(document, &values) != nil {
		return lifecycle.Page{}, lifecycle.ErrUnavailable
	}
	page := lifecycle.Page{Items: make([]lifecycle.Tenant, 0, len(values))}
	if next != nil {
		page.NextCursor = *next
	}
	for _, value := range values {
		page.Items = append(page.Items, value.tenant())
	}
	return page, nil
}

func (value tenantDocument) tenant() lifecycle.Tenant {
	result := lifecycle.Tenant{ID: value.ID, Name: value.Name, Slug: value.Slug,
		Status: value.Status, Revision: value.Revision, SettingsRevision: value.SettingsRevision,
		CreatedAt: value.CreatedAt.UTC(), UpdatedAt: value.UpdatedAt.UTC()}
	if value.OffboardedAt != nil {
		result.OffboardedAt = value.OffboardedAt.UTC()
	}
	if value.PurgedAt != nil {
		result.PurgedAt = value.PurgedAt.UTC()
	}
	return result
}

func (store *Store) Get(ctx context.Context, environmentID, targetID uuid.UUID) (lifecycle.Tenant, error) {
	if !store.usable(ctx) || environmentID == uuid.Nil || targetID == uuid.Nil {
		return lifecycle.Tenant{}, lifecycle.ErrInvalid
	}
	return scanTenant(store.queryRow(ctx, getSQL, environmentID, targetID), store.classify)
}

func nullUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}
