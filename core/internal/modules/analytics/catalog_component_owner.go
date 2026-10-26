package analytics

import (
	"context"
	"fmt"
	"time"

	cataloglifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/cataloglifecycle"
	catalogownerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/catalogowner"
	analyticsforms "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/formsreference"
	catalogownerapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/catalogcomponentowner"
	cataloglegacyapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/cataloglegacyowner"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func NewCatalogComponentOwner() (*catalogownerapp.Owner, error) {
	base, err := catalogownerpostgres.New()
	if err != nil {
		return nil, err
	}
	legacy, err := cataloglegacyapp.New(base, closedFormsReferenceCoordinator{})
	if err != nil {
		return nil, err
	}
	return catalogownerapp.New(legacy, cataloglifecyclepostgres.NewBinder())
}

type closedFormsReferenceCoordinator struct{}

func (closedFormsReferenceCoordinator) BindVersion(context.Context, platformtx.TenantScope,
	analyticsforms.Source, []analyticsforms.Target, time.Time) error {
	return errClosedFormsReferences
}

func (closedFormsReferenceCoordinator) DecommissionSource(context.Context, platformtx.TenantScope,
	analyticsforms.ID, *analyticsforms.ID, time.Time) error {
	return errClosedFormsReferences
}

func (closedFormsReferenceCoordinator) DecommissionComponent(context.Context, platformtx.TenantScope,
	analyticsforms.ID, time.Time) error {
	return errClosedFormsReferences
}

func (closedFormsReferenceCoordinator) CatalogSource(context.Context, platformtx.TenantScope,
	analyticsforms.ID) (analyticsforms.Source, error) {
	return analyticsforms.Source{}, errClosedFormsReferences
}

var errClosedFormsReferences = fmt.Errorf("analytics forms references are unavailable")

var _ componentowner.Provisioner = (*catalogownerapp.Owner)(nil)
var _ analyticsforms.Coordinator = closedFormsReferenceCoordinator{}
