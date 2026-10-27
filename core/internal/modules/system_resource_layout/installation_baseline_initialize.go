package systemresourcelayout

import (
	"context"
	"errors"
	"fmt"
	"time"

	layoutpostgres "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/adapters/postgres"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	"github.com/google/uuid"
)

var ErrInstallationBaselineInvalid = errors.New("system resource installation baseline is invalid")

func InitializeInstallationBaseline(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationBaselineInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		!validInstallationBaseline(input) {
		return ErrInstallationBaselineInvalid
	}
	binding, err := layoutpostgres.BindInstallationStore(layoutpostgres.InstallationDatabase{
		QueryRow: func(ctx context.Context, statement string, arguments ...any) layoutpostgres.InstallationRow {
			return executor.QueryRow(ctx, statement, arguments...)
		},
	})
	if err != nil {
		return fmt.Errorf("bind system resource installation store: %w", err)
	}
	valid, err := binding.Assert(ctx, uuid.UUID(input.OperationID),
		uuid.UUID(input.TenantID), input.SchemaName)
	if err != nil || !valid {
		return fmt.Errorf("assert system resource installation activation: %w", err)
	}
	clock := presentationClock(func() time.Time { return input.ActivatedAt })
	layouts, err := layoutapp.NewLayoutApplication(binding.Store, layoutIDSource(uuid.NewV7), clock)
	if err != nil {
		return fmt.Errorf("create installation layout application: %w", err)
	}
	publisher, err := layoutapp.NewPresentationApplication(binding.Store, clock)
	if err != nil {
		return fmt.Errorf("create installation presentation application: %w", err)
	}
	ids := make([]uuid.UUID, 0, 3)
	for _, resource := range []struct {
		kind    ResourceKind
		catalog layoutapi.PresentationFieldCatalog
	}{{ResourceUser, input.UserCatalog}, {ResourceOrganization, input.OrganizationCatalog},
		{ResourceRole, input.RoleCatalog}} {
		id, initializeErr := initializeInstallationResource(
			ctx, layouts, publisher, input.OperationID, resource.kind, resource.catalog,
		)
		if initializeErr != nil {
			return initializeErr
		}
		ids = append(ids, id)
	}
	_, err = binding.Record(ctx, layoutpostgres.InstallationRecordInput{
		OperationID: uuid.UUID(input.OperationID), TenantID: uuid.UUID(input.TenantID),
		UserRevisionID: ids[0], OrganizationRevisionID: ids[1],
		RoleRevisionID: ids[2], SchemaName: input.SchemaName,
		InitializedAt: input.ActivatedAt,
	})
	if err != nil {
		return fmt.Errorf("record system resource installation baseline: %w", err)
	}
	return nil
}

func initializeInstallationResource(
	ctx context.Context, layouts layoutapi.LayoutAPI,
	publisher *layoutapp.PresentationApplication, operationID [16]byte,
	resource ResourceKind, catalog layoutapi.PresentationFieldCatalog,
) (uuid.UUID, error) {
	definition, err := layoutapp.InstallationDefaultDefinition(resource, catalog)
	if err != nil {
		return uuid.Nil, err
	}
	id := installationBaselineID(operationID, resource)
	_, err = layouts.Create(ctx, layoutapi.CreateInput{
		ID: id.String(), Resource: resource, Definition: definition,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create %s installation layout: %w", resource, err)
	}
	_, err = publisher.Publish(ctx, layoutapi.PublishPresentationInput{
		RevisionID: id.String(), ExpectedRowRevision: 1, Catalog: catalog,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("publish %s installation layout: %w", resource, err)
	}
	return id, nil
}
