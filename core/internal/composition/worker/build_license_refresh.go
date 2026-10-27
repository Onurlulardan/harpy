package worker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildLicenseRefreshWorkload(
	ctx context.Context,
	pool *platformpostgres.Pool,
	settings config.Worker,
	version string,
) (managedAsyncWorkload, bool, error) {
	if ctx == nil || ctx.Err() != nil || pool == nil || version == "" {
		return nil, false, fmt.Errorf("license refresh dependencies are invalid")
	}
	authority := settings.Authority()
	route, err := routeAuthorityWorkload(authority.Mode())
	if err != nil {
		return nil, false, err
	}
	if route == authorityRouteSaaSRegistry {
		return nil, false, nil
	}
	store, err := bindLicenseAuthorityStore(pool)
	if err != nil {
		return nil, false, err
	}
	clockStore, err := billing.NewLocalLicenseClockStore(store.QueryRow)
	if err != nil {
		return nil, false, err
	}
	now, err := billing.ReadLocalLicenseDatabaseClock(ctx, clockStore)
	if err != nil {
		return nil, false, err
	}
	trust, err := billing.ReadLocalLicenseTrustBundle(
		ctx, authority.LicenseTrustBundle().Path().Value(), now,
	)
	if err != nil {
		return nil, false, err
	}
	source, delivery, err := buildLicenseSource(authority)
	if err != nil {
		return nil, false, err
	}
	reconciler, err := billing.NewLocalLicenseReconciler(
		store.LocalLicenseAuthorityStore, source, trust,
		[]billing.LocalLicenseComponentVersion{
			{Component: "core-worker", Version: version},
		},
	)
	if err != nil {
		return nil, false, err
	}
	runner := &licenseRefreshRunner{
		target: func(ctx context.Context) (billing.LocalLicenseRefreshTarget, error) {
			return billing.AcquireLocalLicenseRefreshTarget(
				ctx, store.LocalLicenseAuthorityStore,
			)
		},
		reconcile: func(
			ctx context.Context,
			command billing.LocalLicenseRefreshCommand,
		) error {
			_, reconcileErr := reconciler.Reconcile(ctx, command)
			return reconcileErr
		},
		delivery: delivery,
	}
	workload, err := newLicenseRefreshWorkload(runner)
	return workload, err == nil, err
}

type workerLicenseStore struct {
	billing.LocalLicenseAuthorityStore
	QueryRow func(context.Context, string, ...any) billing.LocalLicenseDatabaseRow
}

func bindLicenseAuthorityStore(
	pool *platformpostgres.Pool,
) (workerLicenseStore, error) {
	return platformpostgres.BindBillingLicenseAuthority(pool, func(
		database platformpostgres.BillingLicenseAuthorityDatabase,
	) (workerLicenseStore, error) {
		query := func(
			ctx context.Context,
			statement string,
			arguments ...any,
		) billing.LocalLicenseDatabaseRow {
			return database.QueryRow(ctx, statement, arguments...)
		}
		store, err := billing.NewLocalLicenseAuthorityStore(query)
		return workerLicenseStore{LocalLicenseAuthorityStore: store, QueryRow: query}, err
	})
}

func buildLicenseSource(
	authority config.WorkerAuthority,
) (billing.LocalLicenseSource, string, error) {
	source := authority.LicenseSource()
	switch authority.Mode() {
	case config.WorkerAuthorityConnected:
		credential, err := billing.ReadLocalLicenseCredential(
			source.CredentialFile().Path().Value(),
		)
		if err != nil {
			return nil, "", err
		}
		defer clear(credential)
		connected, err := billing.NewConnectedLicenseSource(
			source.IssuerURL().Value(), credential, http.DefaultTransport, 15*time.Second,
		)
		return connected, string(billing.LocalLicenseDeliveryConnected), err
	case config.WorkerAuthorityAirGap:
		file, err := billing.NewLocalLicenseSnapshotFile(source.SnapshotFile().Path().Value())
		return file, string(billing.LocalLicenseDeliveryAirGap), err
	case config.WorkerAuthorityDevelopment:
		file, err := billing.NewLocalLicenseSnapshotFile(source.SnapshotFile().Path().Value())
		return file, string(billing.LocalLicenseDeliveryDevelopment), err
	default:
		return nil, "", fmt.Errorf("license source mode is invalid")
	}
}
