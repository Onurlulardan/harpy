package migrator

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/modules/work"
	"github.com/Onurlulardan/harpy/core/internal/platform/cutover"
	cutoverpostgres "github.com/Onurlulardan/harpy/core/internal/platform/cutover/adapters/postgres"
)

func newCutoverApplyRegistry() (cutover.ApplyRegistry, error) {
	probeContract, err := cutoverpostgres.PlatformProbeApplyContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	platformV1Plan, err := cutoverpostgres.PlatformV1EnvironmentPlanContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	tenancyV1Environment, err := tenancy.V1EnvironmentAdoptionContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	iamV1Authority, err := iam.V1ProtectedAuthorityAdoptionContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	billingV1Base, err := billing.V1BaseAdoptionContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	platformV1Completion, err := cutoverpostgres.PlatformV1EnvironmentCompletionContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	billingGlobalContract, err := billing.GlobalCutoverApplyContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	iamContract, err := iam.CutoverApplyContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	billingTenantContract, err := billing.TenantCutoverApplyContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	analyticsContract, err := analytics.CutoverApplyContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	catalogContract, err := catalog.CutoverApplyContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	workContract, err := work.CutoverApplyContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	documentsContract, err := documents.CutoverApplyContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	notificationsContract, err := notifications.CutoverApplyContract()
	if err != nil {
		return cutover.ApplyRegistry{}, err
	}
	return cutover.NewApplyRegistry(
		probeContract,
		platformV1Plan,
		tenancyV1Environment,
		iamV1Authority,
		billingV1Base,
		platformV1Completion,
		billingGlobalContract,
		iamContract,
		billingTenantContract,
		analyticsContract,
		catalogContract,
		workContract,
		documentsContract,
		notificationsContract,
	)
}
