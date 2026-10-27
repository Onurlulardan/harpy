package platformops

import (
	platformopspostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres"
	alerteventpg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/alertevent"
	alertrulepg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/alertrule"
	recoverycontrolpg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/recoverycontrol"
	recoverykitpg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/recoverykit"
	recoverymanifestpg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/recoverymanifest"
	recoveryprotectionpg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/recoveryprotection"
	recoverysetpg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/recoveryset"
	restoreplanpg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/restoreplan"
	tracepg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/trace"
	tracespanpg "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/tracespan"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type runtimeStore struct {
	*platformopspostgres.Store
	*alerteventpg.AlertEventOperations
	*alertrulepg.AlertRuleOperations
	*recoverycontrolpg.RecoveryControlOperations
	*recoverykitpg.RecoveryKitOperations
	*recoverymanifestpg.RecoveryManifestOperations
	*recoveryprotectionpg.RecoveryProtectionOperations
	*recoverysetpg.RecoverySetOperations
	*restoreplanpg.RestorePlanOperations
	*tracepg.TraceOperations
	*tracespanpg.TraceSpanOperations
}

func newRuntimeStore(store *platformopspostgres.Store) *runtimeStore {
	return &runtimeStore{
		Store:                        store,
		AlertEventOperations:         alerteventpg.NewAlertEventOperations(store),
		AlertRuleOperations:          alertrulepg.NewAlertRuleOperations(store),
		RecoveryControlOperations:    recoverycontrolpg.NewRecoveryControlOperations(store),
		RecoveryKitOperations:        recoverykitpg.NewRecoveryKitOperations(store),
		RecoveryManifestOperations:   recoverymanifestpg.NewRecoveryManifestOperations(store),
		RecoveryProtectionOperations: recoveryprotectionpg.NewRecoveryProtectionOperations(store),
		RecoverySetOperations:        recoverysetpg.NewRecoverySetOperations(store),
		RestorePlanOperations:        restoreplanpg.NewPlanOperations(store),
		TraceOperations:              tracepg.NewTraceOperations(store),
		TraceSpanOperations:          tracespanpg.NewTraceSpanOperations(store),
	}
}

var _ ports.RecoveryStore = (*runtimeStore)(nil)
var _ ports.DiagnosticRunner = (*runtimeStore)(nil)
var _ ports.RecoverySetStore = (*runtimeStore)(nil)
var _ ports.RestoreStore = (*runtimeStore)(nil)
var _ ports.EnvironmentIdentityReader = (*runtimeStore)(nil)
var _ ports.RecoveryProtectionStore = (*runtimeStore)(nil)
var _ ports.TraceStore = (*runtimeStore)(nil)
var _ ports.AlertStore = (*runtimeStore)(nil)
