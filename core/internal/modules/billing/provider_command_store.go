package billing

import (
	"time"

	providerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/provider"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func BindProviderCommandStore(
	scope platformtx.TenantScope,
	catalogs platformasync.WriteCatalogs,
) (providerapp.CommandStore, error) {
	authority, err := platformtx.MintOwnerAuthority(scope, "billing")
	if err != nil || !authority.OwnsScope(scope) {
		return nil, providerapp.ErrConfiguration
	}
	effects, err := newProviderCommandEffects(scope, authority, catalogs)
	if err != nil {
		return nil, err
	}
	return providerpostgres.BindCommand(scope, effects)
}

func BindProviderCommandCommands(
	scope platformtx.TenantScope,
	catalogs platformasync.WriteCatalogs,
) (*providerapp.CommandCommands, error) {
	store, err := BindProviderCommandStore(scope, catalogs)
	if err != nil {
		return nil, err
	}
	return providerapp.NewCommandCommands(
		store, uuid.NewV7,
		func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) },
	)
}
