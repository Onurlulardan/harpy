package billing

import (
	"fmt"
	"reflect"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type ErrorSink func(error)

type Dependencies struct {
	Access  AccessVerifier
	Public  billingapi.PublicQueries
	Tenant  billingapi.TenantOperations
	Catalog billingapi.PlatformCatalog
	Reports billingapi.PlatformReports
	Webhook billingapi.WebhookIngress
	Report  ErrorSink
}

type Handler struct {
	access  AccessVerifier
	public  billingapi.PublicQueries
	tenant  billingapi.TenantOperations
	catalog billingapi.PlatformCatalog
	reports billingapi.PlatformReports
	webhook billingapi.WebhookIngress
	query   billingapi.QueryDecoder
	report  ErrorSink
}

func New(dependencies Dependencies) (*Handler, error) {
	if nilPort(dependencies.Access) || nilPort(dependencies.Public) ||
		nilPort(dependencies.Tenant) || nilPort(dependencies.Catalog) ||
		nilPort(dependencies.Reports) || nilPort(dependencies.Webhook) ||
		dependencies.Report == nil {
		return nil, fmt.Errorf("billing HTTP dependencies are required")
	}
	return &Handler{
		access: dependencies.Access, public: dependencies.Public,
		tenant: dependencies.Tenant, catalog: dependencies.Catalog,
		reports: dependencies.Reports, webhook: dependencies.Webhook,
		query: billingapi.NewQueryDecoder(), report: dependencies.Report,
	}, nil
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
