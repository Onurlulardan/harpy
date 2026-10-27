package iam

import (
	setuppostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/initialadminsetup"
	setupapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/initialadminsetup"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type InitialAdminSetupPreparer = setupapp.Preparer
type InitialAdminSetupPrepareInput = setupapp.PrepareInput

func NewInitialAdminSetupPreparer(
	pool setuppostgres.RuntimePool,
	settings config.IAMAPI,
	delivery *DeliverySealer,
) (*setupapp.Preparer, error) {
	store, err := setuppostgres.NewRuntimeStore(pool)
	if err != nil {
		return nil, err
	}
	keys, err := settings.OneTimeTokenKeyring()
	if err != nil {
		return nil, err
	}
	return setupapp.NewPreparer(setupapp.Dependencies{
		Store: store, Keys: keys, Delivery: delivery,
	})
}
