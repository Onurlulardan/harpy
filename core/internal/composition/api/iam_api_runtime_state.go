package api

import (
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	systemresourcelayouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type iamAPIRuntime struct {
	identityRuntime            *iammodule.IdentityRuntime
	profileCommands            *iammodule.ProfileCommands
	profileBackfills           *iammodule.ProfileBackfillPublicationRuntime
	profileRecoveries          *iammodule.ProfileBackfillRecoveryRuntime
	rbacRuntime                *iammodule.RBACRuntime
	restQueryRuntime           *iammodule.IAMRESTQueryRuntime
	security                   *iamHTTPSecurity
	billingAccess              *iammodule.BillingAccessRuntime
	usageAccess                *iammodule.UsageAccessRuntime
	supportAccess              *iammodule.SupportAccessRuntime
	auditAccess                *iammodule.AuditAccessRuntime
	tenantSettingsAccess       *iammodule.TenantSettingsAccessRuntime
	platformAccess             *iammodule.PlatformAccessRuntime
	customerProvisioningAccess *iammodule.CustomerProvisioningAccessRuntime
	formsRecordAccess          *iammodule.FormsRecordAccessRuntime
	portalAccess               *iammodule.PortalAccessRuntime
	formsLifecycleAccess       *iammodule.FormsLifecycleAccessRuntime
	workflowAccess             *iammodule.WorkflowAccessRuntime
	analyticsAccess            *iammodule.AnalyticsAccessRuntime
	catalogAccess              *iammodule.CatalogAccessRuntime
	documentsAccess            *iammodule.DocumentsAccessRuntime
	documentsArchiveAccess     *iammodule.DocumentsArchiveAccessRuntime
	notificationsAccess        *iammodule.NotificationsAccessRuntime
	workAccess                 *iammodule.WorkAccessRuntime
	analyticsPrincipals        *iammodule.AnalyticsPrincipalPermissionRuntime
	servicePrincipals          *iammodule.ServicePrincipalRuntime
	humanTaskAccess            *iammodule.HumanTaskAccessRuntime
	systemLayoutAccess         *iammodule.SystemLayoutAccessRuntime
	authApplication            iammodule.AuthHTTPApplication
	identityApplication        iammodule.IdentityHTTPApplication
	systemResources            *systemresources.Application
	systemLayouts              *systemresourcelayouts.Application
	profileHTTP                iammodule.ProfileHTTPApplication
	invitationRuntime          *iammodule.InvitationRuntime
	invitationHTTP             iammodule.InvitationHTTPApplication
	permissionRequests         *iammodule.PermissionRequestRuntime
	permissionHTTP             iammodule.PermissionRequestHTTPApplication
	lifecycleRuntime           *iammodule.LifecycleRuntime
	lifecycleHTTP              iammodule.LifecycleHTTPApplication
	rbacHTTP                   iammodule.RBACHTTPApplication
	passwordBlocklists         *iammodule.PasswordBlocklists
	deliverySealer             *iammodule.DeliverySealer
	totpDecryptor              iamTOTPDecryptor
	totpSealer                 *iammodule.TOTPSealer
	actorBridge                *iammodule.ActorBridgeRuntime
	systemTransferHTTP         *iammodule.SystemTransferHTTP
	postRestore                *iammodule.PostRestoreRuntime
	coordinator                *platformtx.Coordinator
	identityChanges            iammodule.IdentityMutationSink
	routes                     [][]httpedge.Route
}

func (runtime *iamAPIRuntime) Close() {
	if runtime == nil {
		return
	}
	if runtime.security != nil {
		runtime.security.Close()
	}
	if runtime.actorBridge != nil {
		runtime.actorBridge.Close()
	}
	if runtime.totpDecryptor != nil {
		_ = runtime.totpDecryptor.Close()
	}
	if runtime.totpSealer != nil {
		_ = runtime.totpSealer.Close()
	}
	if runtime.deliverySealer != nil {
		_ = runtime.deliverySealer.Close()
	}
	if runtime.passwordBlocklists != nil {
		_ = runtime.passwordBlocklists.Close()
	}
}

func (runtime *iamAPIRuntime) actorBridgeReady() bool {
	return runtime != nil && runtime.actorBridge != nil && runtime.actorBridge.Ready() &&
		runtime.systemTransferHTTP != nil && runtime.systemTransferHTTP.Ready()
}
