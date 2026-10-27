package postgres

import _ "embed"

import cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
import platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"

//go:embed sql/observe_session.sql
var observeSessionSQL string

//go:embed sql/observe_role_lease.sql
var observeRoleLeaseSQL string

//go:embed sql/observe_sealed_role.sql
var observeSealedRoleSQL string

//go:embed sql/prepare_batch.sql
var prepareBatchSQL string

//go:embed sql/apply_cutover_probe_v1.sql
var applyCutoverProbeV1SQL string

//go:embed sql/apply_v1_environment_adoption.sql
var applyV1EnvironmentAdoptionSQL string

//go:embed sql/complete_v1_environment_adoption.sql
var completeV1EnvironmentAdoptionSQL string

//go:embed sql/finalize_operation.sql
var finalizeOperationSQL string

//go:embed sql/observe_finalizer_procedure.sql
var observeFinalizerProcedureSQL string

func PlatformProbeApplyContract() (platformcutover.ApplyContract, error) {
	return platformcutover.NewApplyContract(
		"platform.cutover_probe.v1", applyCutoverProbeV1SQL,
	)
}

func PlatformV1EnvironmentPlanContract() (platformcutover.ApplyContract, error) {
	return platformcutover.NewApplyContract(
		cutoverv1.PlatformV1EnvironmentPlanContractID,
		applyV1EnvironmentAdoptionSQL,
	)
}

func PlatformV1EnvironmentCompletionContract() (platformcutover.ApplyContract, error) {
	return platformcutover.NewApplyContract(
		cutoverv1.PlatformV1EnvironmentCompletionContractID,
		completeV1EnvironmentAdoptionSQL,
	)
}
