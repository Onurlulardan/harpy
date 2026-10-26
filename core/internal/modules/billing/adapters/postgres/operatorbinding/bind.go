package operatorbinding

import platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"

type Pool interface {
	PostgresOwnerBindingPool() *platformpostgres.Pool
}

type offeringDatabase struct {
	bound platformpostgres.BillingOperatorOfferingDatabase
}

type saasBaseDatabase struct {
	bound platformpostgres.BillingSaaSBaseDatabase
}

func WithInstallationOffering[T any](
	pool Pool,
	factory func(OfferingDatabase) (T, error),
) (T, error) {
	return platformpostgres.BindBillingOperatorOffering(
		pool.PostgresOwnerBindingPool(),
		func(bound platformpostgres.BillingOperatorOfferingDatabase) (T, error) {
			return factory(offeringDatabase{bound: bound})
		},
	)
}

func WithSaaSBase[T any](
	pool Pool,
	factory func(SaaSBaseDatabase) (T, error),
) (T, error) {
	return platformpostgres.BindBillingSaaSBase(
		pool.PostgresOwnerBindingPool(),
		func(bound platformpostgres.BillingSaaSBaseDatabase) (T, error) {
			return factory(saasBaseDatabase{bound: bound})
		},
	)
}
