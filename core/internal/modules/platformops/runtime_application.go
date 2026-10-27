package platformops

import (
	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	platformopsapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application"
	alerteventapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/alertevent"
	alertruleapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/alertrule"
	recoverycontrolapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoverycontrol"
	recoverykitapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoverykit"
	recoverymanifestapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoverymanifest"
	recoveryprotectionapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryprotection"
	recoverysetapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryset"
	restoreplanapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/restoreplan"
	systemhealthapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/systemhealth"
	traceapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/trace"
	tracespanapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/tracespan"
)

type runtimeOperations struct {
	*platformopsapp.OperationsService
	*alerteventapp.AlertEventOperations
	*alertruleapp.AlertRuleOperations
	*systemhealthapp.SystemHealthOperations
	*traceapp.TraceOperations
	*tracespanapp.TraceSpanOperations
}

var _ platformopsapi.HealthReader = (*runtimeOperations)(nil)
var _ platformopsapi.TraceOperations = (*runtimeOperations)(nil)
var _ platformopsapi.AlertOperations = (*runtimeOperations)(nil)
var _ platformopsapi.RecoverySets = (*runtimeRecovery)(nil)
var _ platformopsapi.Restores = (*runtimeRecovery)(nil)
var _ platformopsapi.RecoveryProtectionService = (*runtimeProtection)(nil)

func newRuntimeOperations(service *platformopsapp.OperationsService) *runtimeOperations {
	return &runtimeOperations{OperationsService: service,
		AlertEventOperations:   alerteventapp.NewAlertEventOperations(service),
		AlertRuleOperations:    alertruleapp.NewAlertRuleOperations(service),
		SystemHealthOperations: systemhealthapp.NewSystemHealthOperations(service),
		TraceOperations:        traceapp.NewTraceOperations(service),
		TraceSpanOperations:    tracespanapp.NewTraceSpanOperations(service)}
}

type runtimeRecovery struct {
	*platformopsapp.RecoveryService
	*recoverymanifestapp.RecoveryManifestOperations
	*recoverysetapp.RecoverySetOperations
	*restoreplanapp.RestorePlanOperations
}

func newRuntimeRecovery(service *platformopsapp.RecoveryService) *runtimeRecovery {
	return &runtimeRecovery{RecoveryService: service,
		RecoveryManifestOperations: recoverymanifestapp.NewRecoveryManifestOperations(service),
		RecoverySetOperations:      recoverysetapp.NewRecoverySetOperations(service),
		RestorePlanOperations:      restoreplanapp.NewPlanOperations(service)}
}

type runtimeProtection struct {
	*platformopsapp.RecoveryProtectionService
	*recoverycontrolapp.RecoveryControlOperations
	*recoverykitapp.RecoveryKitOperations
	*recoveryprotectionapp.RecoveryProtectionOperations
}

func newRuntimeProtection(
	service *platformopsapp.RecoveryProtectionService,
) *runtimeProtection {
	return &runtimeProtection{RecoveryProtectionService: service,
		RecoveryControlOperations:    recoverycontrolapp.NewRecoveryControlOperations(service),
		RecoveryKitOperations:        recoverykitapp.NewRecoveryKitOperations(service),
		RecoveryProtectionOperations: recoveryprotectionapp.NewRecoveryProtectionOperations(service)}
}
