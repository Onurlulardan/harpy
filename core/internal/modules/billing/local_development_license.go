package billing

import (
	"context"
	"errors"
	"net/http"
	"time"

	licensebootstrapv1 "github.com/Onurlulardan/harpy/contracts/licensebootstrap/v1"
	licenseauthoritysource "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/licenseauthoritysource"
	licenseauthoritypostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/licenseauthority"
	licenseclockpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/licenseclock"
	licenseofferingpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/licenseoffering"
	licenseauthorityapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	licenseclockapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseclock"
	licenseofferingapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseoffering"
	"github.com/google/uuid"
)

type LocalLicenseTrustBundle = licenseauthorityapp.TrustBundle
type LocalLicenseVerifiedSnapshot = licenseauthorityapp.VerifiedSnapshot
type LocalLicenseComponentVersion = licenseauthorityapp.ComponentVersion
type LocalLicenseRefreshCommand = licenseauthorityapp.RefreshCommand
type LocalLicenseIssueRequest = licenseauthorityapp.IssueRequest
type LocalLicenseBinding = licenseauthorityapp.Binding
type LocalLicenseOfferingPlan = licenseofferingapp.Plan
type LocalLicenseOfferingDocument = licenseofferingapp.Document
type LocalLicenseAuthorityStore = licenseauthoritypostgres.Store
type LocalLicenseClockStore = licenseclockpostgres.Store
type LocalLicenseSource = licenseauthorityapp.Source
type LocalLicenseRefreshTarget = licenseauthorityapp.RefreshTarget
type LocalLicenseRestoreBinding = licenseauthorityapp.RestoreBinding
type LocalRestoredLicenseCoordinator = licenseauthorityapp.RestoredLicenseCoordinator
type LocalRestoredLicenseReceipt = licenseauthorityapp.RestoredLicenseReceipt
type UUID = uuid.UUID

const (
	LocalLicenseDeliveryConnected   = licenseauthorityapp.DeliveryConnected
	LocalLicenseDeliveryAirGap      = licenseauthorityapp.DeliveryAirGap
	LocalLicenseDeliveryDevelopment = licenseauthorityapp.DeliveryDevelopment
)

type LocalLicenseDatabaseRow interface{ Scan(...any) error }

func ReadLocalLicenseTrustBundle(
	ctx context.Context,
	path string,
	now time.Time,
) (LocalLicenseTrustBundle, error) {
	return licenseauthoritysource.ReadTrustBundle(ctx, path, now)
}

func NewLocalLicenseSnapshotFile(
	path string,
) (*licenseauthoritysource.SnapshotFile, error) {
	return licenseauthoritysource.NewSnapshotFile(path)
}

func NewConnectedLicenseSource(
	baseURL string,
	credential []byte,
	transport http.RoundTripper,
	timeout time.Duration,
) (*licenseauthoritysource.Connected, error) {
	return licenseauthoritysource.NewConnected(
		baseURL, credential, transport, timeout,
	)
}

func ReadLocalLicenseCredential(path string) ([]byte, error) {
	return licenseauthoritysource.ReadCredential(path)
}

func NewLocalLicenseAuthorityStore(
	queryRow func(context.Context, string, ...any) LocalLicenseDatabaseRow,
) (licenseauthoritypostgres.Store, error) {
	return licenseauthoritypostgres.NewStore(licenseauthoritypostgres.DatabaseAccess{
		QueryRow: func(ctx context.Context, statement string, arguments ...any) licenseauthoritypostgres.Row {
			return queryRow(ctx, statement, arguments...)
		},
	})
}

func ReadLocalLicenseDatabaseClock(
	ctx context.Context,
	store LocalLicenseClockStore,
) (time.Time, error) {
	return licenseclockapp.Read(ctx, store)
}

func NewLocalLicenseClockStore(
	queryRow func(context.Context, string, ...any) LocalLicenseDatabaseRow,
) (licenseclockpostgres.Store, error) {
	return licenseclockpostgres.NewStore(licenseclockpostgres.DatabaseAccess{
		QueryRow: func(ctx context.Context, statement string, arguments ...any) licenseclockpostgres.Row {
			return queryRow(ctx, statement, arguments...)
		},
	})
}

func NewLocalLicenseOfferingStore(
	queryRow func(context.Context, string, ...any) LocalLicenseDatabaseRow,
) (licenseofferingapp.Store, error) {
	return licenseofferingpostgres.NewStore(licenseofferingpostgres.DatabaseAccess{
		QueryRow: func(ctx context.Context, statement string, arguments ...any) licenseofferingpostgres.Row {
			return queryRow(ctx, statement, arguments...)
		},
	})
}

func NewLocalLicenseReconciler(
	store licenseauthorityapp.Store,
	source licenseauthorityapp.Source,
	trust licenseauthorityapp.KeyResolver,
	versions []LocalLicenseComponentVersion,
) (*licenseauthorityapp.Reconciler, error) {
	return licenseauthorityapp.NewLicenseReconciler(store, source, trust, versions)
}

func AcquireLocalLicenseRefreshTarget(
	ctx context.Context,
	store licenseauthorityapp.RefreshTargetStore,
) (LocalLicenseRefreshTarget, error) {
	return licenseauthorityapp.AcquireRefreshTarget(ctx, store)
}

func LoadLocalLicenseRestoreBinding(
	ctx context.Context,
	store licenseauthorityapp.RestoreBindingStore,
	environmentID UUID,
) (LocalLicenseRestoreBinding, error) {
	return licenseauthorityapp.ResolveRestoreBinding(ctx, store, environmentID)
}

func IssueLocalLicenseSnapshot(
	ctx context.Context,
	source LocalLicenseSource,
	request LocalLicenseIssueRequest,
) ([]byte, error) {
	if source == nil {
		return nil, licenseauthorityapp.ErrIssueRequest
	}
	return source.IssueSnapshot(ctx, request)
}

func VerifyLocalLicenseSnapshot(
	ctx context.Context,
	document []byte,
	binding LocalLicenseBinding,
	now time.Time,
	trust LocalLicenseTrustBundle,
) (LocalLicenseVerifiedSnapshot, error) {
	return licenseauthorityapp.Verify(ctx, document, binding, now, trust)
}

func NewLocalRestoredLicenseCoordinator(
	store licenseauthorityapp.RestoredLicenseStore,
	trust licenseauthorityapp.KeyResolver,
) (*LocalRestoredLicenseCoordinator, error) {
	return licenseauthorityapp.NewRestoredLicenseCoordinator(store, trust)
}

func LocalLicenseRefreshRetryable(err error) bool {
	return errors.Is(err, licenseauthorityapp.ErrRefreshRetryable)
}

func LocalLicenseRefreshFenced(err error) bool {
	return errors.Is(err, licenseauthorityapp.ErrRefreshFenced)
}

func NewLocalLicenseSnapshotChecker(
	trust licenseauthorityapp.KeyResolver,
) (*licenseauthorityapp.SnapshotChecker, error) {
	return licenseauthorityapp.NewSnapshotChecker(trust)
}

func BuildLocalLicenseOfferingDocument(
	plan LocalLicenseOfferingPlan,
	snapshot LocalLicenseVerifiedSnapshot,
) (LocalLicenseOfferingDocument, error) {
	return licenseofferingapp.BuildDocument(plan, snapshot)
}

func NewLocalLicenseOfferingPublisher(
	store licenseofferingapp.Store,
	checker *licenseauthorityapp.SnapshotChecker,
) (*licenseofferingapp.Publisher, error) {
	return licenseofferingapp.NewPublisher(store, checker)
}

func LocalLicenseBootstrapManifestSHA256() string {
	return licensebootstrapv1.ManifestSHA256()
}

func IsUUIDv7(value UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}
