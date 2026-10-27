package operator

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type billingDatabase struct {
	inner platformpostgres.BillingOperatorOfferingDatabase
}

func bindBillingDatabase(pool *platformpostgres.Pool) (*billingDatabase, error) {
	return platformpostgres.BindBillingOperatorOffering(pool,
		func(database platformpostgres.BillingOperatorOfferingDatabase) (
			*billingDatabase, error,
		) {
			return &billingDatabase{inner: database}, nil
		})
}

func (database *billingDatabase) clock(ctx context.Context) (databaseClock, error) {
	store, err := billing.NewLocalLicenseClockStore(database.queryRow)
	if err != nil {
		return databaseClock{}, err
	}
	now, err := billing.ReadLocalLicenseDatabaseClock(ctx, store)
	return databaseClock{now: now}, err
}

func (database *billingDatabase) licenseStore() (billing.LocalLicenseAuthorityStore, error) {
	return billing.NewLocalLicenseAuthorityStore(database.queryRow)
}

func (database *billingDatabase) queryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) billing.LocalLicenseDatabaseRow {
	return database.inner.QueryRow(ctx, statement, arguments...)
}
