package api

import (
	"fmt"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *iamAPIRuntime) buildHTTPAccess(
	pool *postgres.Pool,
	coordinator *platformtx.Coordinator,
	catalogs platformasync.WriteCatalogs,
	environmentID [16]byte,
) error {
	var err error
	runtime.billingAccess, err = iammodule.NewBillingAccessRuntime(
		pool, coordinator, runtime.security, time.Now,
	)
	if err != nil {
		return fmt.Errorf("build IAM Billing access runtime: %w", err)
	}
	runtime.usageAccess, err = iammodule.NewUsageAccessRuntime(
		runtime.billingAccess, environmentID,
	)
	if err != nil {
		return fmt.Errorf("build IAM Usage access runtime: %w", err)
	}
	runtime.supportAccess, err = iammodule.NewSupportAccessRuntime(runtime.billingAccess)
	if err != nil {
		return fmt.Errorf("build IAM Support access runtime: %w", err)
	}
	runtime.auditAccess, err = iammodule.NewAuditAccessRuntime(runtime.billingAccess)
	if err != nil {
		return fmt.Errorf("build IAM Audit access runtime: %w", err)
	}
	runtime.tenantSettingsAccess, err = iammodule.NewTenantSettingsAccessRuntime(
		runtime.billingAccess)
	if err != nil {
		return fmt.Errorf("build IAM tenant settings access runtime: %w", err)
	}
	runtime.platformAccess, err = iammodule.NewPlatformAccessRuntime(runtime.billingAccess)
	if err != nil {
		return fmt.Errorf("build IAM platform access runtime: %w", err)
	}
	runtime.customerProvisioningAccess, err =
		iammodule.NewCustomerProvisioningAccessRuntime(runtime.billingAccess)
	if err != nil {
		return fmt.Errorf("build IAM customer provisioning access runtime: %w", err)
	}
	runtime.formsRecordAccess, err = iammodule.NewFormsRecordAccessRuntime(
		pool, coordinator, runtime.security, time.Now,
	)
	if err != nil {
		return fmt.Errorf("build IAM record access runtime: %w", err)
	}
	runtime.portalAccess, err = iammodule.NewPortalAccessRuntime(
		pool, coordinator, runtime.security, time.Now,
	)
	if err != nil {
		return fmt.Errorf("build IAM portal access runtime: %w", err)
	}
	runtime.analyticsAccess, err = iammodule.NewAnalyticsAccessRuntime(
		runtime.formsRecordAccess,
	)
	if err != nil {
		return fmt.Errorf("build IAM Analytics access runtime: %w", err)
	}
	runtime.catalogAccess, err = iammodule.NewCatalogAccessRuntime(
		runtime.formsRecordAccess,
	)
	if err != nil {
		return fmt.Errorf("build IAM Catalog access runtime: %w", err)
	}
	runtime.documentsAccess, err = iammodule.NewDocumentsAccessRuntime(
		runtime.formsRecordAccess,
	)
	if err != nil {
		return fmt.Errorf("build IAM Documents access runtime: %w", err)
	}
	runtime.documentsArchiveAccess, err = iammodule.NewDocumentsArchiveAccessRuntime(
		pool, coordinator,
	)
	if err != nil {
		return fmt.Errorf("build IAM Documents archive access runtime: %w", err)
	}
	runtime.notificationsAccess, err = iammodule.NewNotificationsAccessRuntime(
		runtime.formsRecordAccess,
	)
	if err != nil {
		return fmt.Errorf("build IAM Notifications access runtime: %w", err)
	}
	runtime.workAccess, err = iammodule.NewWorkAccessRuntime(runtime.formsRecordAccess)
	if err != nil {
		return fmt.Errorf("build IAM Work access runtime: %w", err)
	}
	runtime.analyticsPrincipals, err = iammodule.NewAnalyticsPrincipalPermissionRuntime()
	if err != nil {
		return fmt.Errorf("build IAM Analytics principal permission runtime: %w", err)
	}
	_ = catalogs
	runtime.systemLayoutAccess, err = iammodule.NewSystemLayoutAccessRuntime(
		pool, coordinator, runtime.security, time.Now,
	)
	if err != nil {
		return fmt.Errorf("build IAM system layout access runtime: %w", err)
	}
	return nil
}
