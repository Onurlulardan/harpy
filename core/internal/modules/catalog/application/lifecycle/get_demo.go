package lifecycle

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) GetDemoStatus(
	ctx context.Context,
	scope platformtx.TenantScope,
	installationID uuid.UUID,
) (DemoStatus, error) {
	if installationID.Version() != 7 {
		return DemoStatus{}, ErrInvalid
	}
	store, demoStore, _, err := application.bindDemoStore(
		ctx, scope, ScopeSnapshot, operationDemoStatus, ActionRead,
	)
	if err != nil {
		return DemoStatus{}, err
	}
	reader, ok := store.(InstallationByIDStore)
	if !ok || reader == nil {
		return DemoStatus{}, ErrUnavailable
	}
	if _, err = reader.InstallationByID(ctx, installationID); err != nil {
		return DemoStatus{}, ErrNotFound
	}
	components, err := demoStore.DemoComponents(ctx, installationID)
	if err != nil {
		return DemoStatus{}, err
	}
	return compileDemoStatus(installationID, components)
}

func compileDemoStatus(
	installationID uuid.UUID,
	components []DemoComponent,
) (DemoStatus, error) {
	if len(components) == 0 {
		return DemoStatus{InstallationID: installationID,
			State: DemoStateNotIncluded}, nil
	}
	result := DemoStatus{InstallationID: installationID,
		State: components[0].State, Components: make([]DemoComponent, len(components))}
	for index, component := range components {
		if !validDemoComponent(component, installationID) || component.State != result.State {
			return DemoStatus{}, ErrUnavailable
		}
		result.Components[index] = component.Clone()
		if component.UpdatedAt.After(result.UpdatedAt) {
			result.UpdatedAt = component.UpdatedAt
		}
	}
	return result, nil
}

func validDemoComponent(value DemoComponent, installationID uuid.UUID) bool {
	return value.InstallationID == installationID && value.PackageID != uuid.Nil &&
		value.ComponentID.Version() == 7 && value.ReleaseID.Version() == 7 &&
		value.SourceRunID.Version() == 7 && value.PackageVersion != "" &&
		sha256Pattern.MatchString(value.ContentSHA256) && len(value.Document) > 0 &&
		(value.State == DemoStateInstalled || value.State == DemoStateAbsent) &&
		value.Revision > 0 && !value.UpdatedAt.IsZero() &&
		value.UpdatedAt.Location() == time.UTC && value.UpdatedAt.Nanosecond()%1000 == 0
}
