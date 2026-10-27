package billingv1

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
	"github.com/google/uuid"
)

type Dependencies struct {
	Access       AccessVerifier
	Public       billingapi.PublicQueries
	Tenant       billingapi.TenantOperations
	Catalog      billingapi.PlatformCatalog
	Identity     billingapi.PublicLegacyIdentityResolver
	Reports      billingapi.PlatformReports
	Retirement   contractretirement.RuntimePort
	Availability AvailabilityPolicy
	Clock        func() time.Time
	NextID       func() (uuid.UUID, error)
	CouponCode   func() (string, error)
}

type Handler struct {
	access           AccessVerifier
	public           billingapi.PublicQueries
	tenant           billingapi.TenantOperations
	catalog          billingapi.PlatformCatalog
	identity         billingapi.PublicLegacyIdentityResolver
	reports          billingapi.PlatformReports
	retirement       contractretirement.RuntimePort
	availability     AvailabilityPolicy
	clock            func() time.Time
	nextID           func() (uuid.UUID, error)
	couponCodeSource func() (string, error)
	query            billingapi.QueryDecoder
	manifest         contractretirement.Batch
	keys             map[string]contractretirement.Batch
	admitted         atomic.Bool
}

func New(dependencies Dependencies) (*Handler, error) {
	if nilPort(dependencies.Access) || nilPort(dependencies.Public) ||
		nilPort(dependencies.Tenant) || nilPort(dependencies.Catalog) ||
		nilPort(dependencies.Identity) ||
		nilPort(dependencies.Reports) || nilPort(dependencies.Retirement) ||
		nilPort(dependencies.Availability) ||
		dependencies.Clock == nil || dependencies.NextID == nil ||
		dependencies.CouponCode == nil {
		return nil, fmt.Errorf("Billing V1 compatibility dependencies are required")
	}
	manifest, err := RetirementManifest()
	if err != nil {
		return nil, err
	}
	handler := &Handler{
		access: dependencies.Access, public: dependencies.Public,
		tenant: dependencies.Tenant, catalog: dependencies.Catalog,
		identity: dependencies.Identity,
		reports:  dependencies.Reports, retirement: dependencies.Retirement,
		availability: dependencies.Availability,
		clock:        dependencies.Clock, nextID: dependencies.NextID,
		couponCodeSource: dependencies.CouponCode,
		query:            billingapi.NewQueryDecoder(), manifest: manifest,
		keys: make(map[string]contractretirement.Batch, len(Routes())),
	}
	identity, _ := RetirementIdentity()
	for _, route := range Routes() {
		key, keyErr := RetirementKey(route)
		if keyErr != nil {
			return nil, keyErr
		}
		handler.keys[route.ID], err = contractretirement.NewBatch(
			identity, []contractretirement.Key{key},
		)
		if err != nil {
			return nil, err
		}
	}
	return handler, nil
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
