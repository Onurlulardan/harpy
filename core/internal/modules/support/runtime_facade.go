package support

import (
	"fmt"

	supportpostgres "github.com/Onurlulardan/harpy/core/internal/modules/support/adapters/postgres"
	contactpostgres "github.com/Onurlulardan/harpy/core/internal/modules/support/adapters/postgres/contact"
	platformbugpostgres "github.com/Onurlulardan/harpy/core/internal/modules/support/adapters/postgres/platformbug"
	reporterbugpostgres "github.com/Onurlulardan/harpy/core/internal/modules/support/adapters/postgres/reporterbug"
	tenantbugpostgres "github.com/Onurlulardan/harpy/core/internal/modules/support/adapters/postgres/tenantbug"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	contactapplication "github.com/Onurlulardan/harpy/core/internal/modules/support/application/contact"
	platformbugapplication "github.com/Onurlulardan/harpy/core/internal/modules/support/application/platformbug"
	reporterbugapplication "github.com/Onurlulardan/harpy/core/internal/modules/support/application/reporterbug"
	tenantbugapplication "github.com/Onurlulardan/harpy/core/internal/modules/support/application/tenantbug"
)

type RuntimeDatabase = supportpostgres.Database
type RuntimeRow = supportpostgres.Row
type RuntimeTransactionMode = supportpostgres.TransactionMode
type RuntimeTransactionRunner = supportpostgres.TransactionRunner
type RuntimeClock = application.Clock
type RuntimeIDGenerator = application.IDGenerator
type RuntimeID = application.ID

type RuntimeHTTPDependencies struct {
	Run         RuntimeTransactionRunner
	RowNotFound func(error) bool
	Clock       RuntimeClock
	IDs         RuntimeIDGenerator
	HTTP        HTTPHandlerDependencies
}

func NewRuntimeHTTPHandler(dependencies RuntimeHTTPDependencies) (*HTTPHandler, error) {
	driver, err := supportpostgres.NewStoreForProfile(dependencies.Run, dependencies.RowNotFound)
	if err != nil {
		return nil, err
	}
	contactPersistence, err := contactpostgres.NewStore(driver)
	if err != nil {
		return nil, err
	}
	reporterPersistence, err := reporterbugpostgres.NewStore(driver)
	if err != nil {
		return nil, err
	}
	tenantPersistence, err := tenantbugpostgres.NewStore(driver)
	if err != nil {
		return nil, err
	}
	platformPersistence, err := platformbugpostgres.NewStore(driver)
	if err != nil {
		return nil, err
	}
	store, err := supportpostgres.NewApplicationStore(contactPersistence, reporterPersistence,
		tenantPersistence, platformPersistence)
	if err != nil {
		return nil, err
	}
	kernel, err := application.NewService(store, dependencies.Clock, dependencies.IDs)
	if err != nil {
		return nil, err
	}
	applications, err := newRuntimeApplications(kernel)
	if err != nil {
		return nil, err
	}
	dependencies.HTTP.Applications = applications
	return NewHTTPHandler(dependencies.HTTP)
}

func newRuntimeApplications(kernel *application.Service) (HTTPApplications, error) {
	contact, err := contactapplication.NewService(kernel)
	if err != nil {
		return HTTPApplications{}, err
	}
	reporter, err := reporterbugapplication.NewService(kernel)
	if err != nil {
		return HTTPApplications{}, err
	}
	tenant, err := tenantbugapplication.NewService(kernel)
	if err != nil {
		return HTTPApplications{}, err
	}
	platform, err := platformbugapplication.NewService(kernel)
	if err != nil {
		return HTTPApplications{}, err
	}
	applications := HTTPApplications{Contact: contact, Reporter: reporter,
		Tenant: tenant, Platform: platform}
	if applications.Contact == nil || applications.Reporter == nil ||
		applications.Tenant == nil || applications.Platform == nil {
		return HTTPApplications{}, fmt.Errorf("support runtime applications are unavailable")
	}
	return applications, nil
}
