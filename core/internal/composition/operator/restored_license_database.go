package operator

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	platformops "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type restoredLicenseDatabase struct {
	billing billing.LocalLicenseAuthorityStore
	clock   billing.LocalLicenseClockStore
	grants  platformops.RestoredLicenseGrantStore
}

func bindRestoredLicenseDatabase(
	pool *platformpostgres.Pool,
) (*restoredLicenseDatabase, error) {
	var query func(context.Context, string, ...any) billing.LocalLicenseDatabaseRow
	billingStore, err := platformpostgres.BindBillingLicenseAuthority(
		pool,
		func(database platformpostgres.BillingLicenseAuthorityDatabase) (
			billing.LocalLicenseAuthorityStore, error,
		) {
			query = func(
				ctx context.Context,
				statement string,
				arguments ...any,
			) billing.LocalLicenseDatabaseRow {
				return database.QueryRow(ctx, statement, arguments...)
			}
			return billing.NewLocalLicenseAuthorityStore(query)
		},
	)
	if err != nil {
		return nil, err
	}
	clockStore, err := billing.NewLocalLicenseClockStore(query)
	if err != nil {
		return nil, err
	}
	grantStore, err := platformops.NewRestoredLicenseGrantStore(pool)
	if err != nil {
		return nil, err
	}
	return &restoredLicenseDatabase{
		billing: billingStore, clock: clockStore, grants: grantStore,
	}, nil
}
