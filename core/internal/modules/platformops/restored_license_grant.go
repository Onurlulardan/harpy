package platformops

import (
	"context"
	"time"

	grantpostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/restoredlicense"
	grantapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/restoredlicense"
)

type RestoredLicenseGrantInput = grantapp.Input
type RestoredLicenseGrantResult = grantapp.Result
type RestoredLicenseGrantStore = grantapp.Store
type RestoredLicenseGrant = grantapp.Grant
type RestoredLicenseScope = grantapp.Scope

func NewRestoredLicenseGrantStore(
	database RuntimeDatabase,
) (grantapp.Store, error) {
	return grantpostgres.NewStore(database)
}

func MintRestoredLicenseGrant(
	ctx context.Context,
	store grantapp.Store,
	input RestoredLicenseGrantInput,
) (RestoredLicenseGrantResult, error) {
	return grantapp.Reconcile(ctx, store, input)
}

func RunRestoredLicenseGrant(
	grant RestoredLicenseGrant,
	databaseNow time.Time,
	callback func(RestoredLicenseScope) error,
) error {
	return grantapp.Run(grant, databaseNow, callback)
}
