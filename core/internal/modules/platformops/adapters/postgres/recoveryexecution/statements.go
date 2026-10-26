package recoveryexecution

import _ "embed"

var (
	//go:embed sql/platform/claim_next_recovery_set.sql
	claimNextRecoverySetSQL string
	//go:embed sql/platform/heartbeat_recovery_execution.sql
	heartbeatRecoveryExecutionSQL string
	//go:embed sql/platform/attest_documents_hold.sql
	attestDocumentsHoldSQL string
	//go:embed sql/platform/record_recovery_component.sql
	recordRecoveryComponentSQL string
	//go:embed sql/platform/put_recovery_components.sql
	putRecoveryComponentsSQL string
	//go:embed sql/platform/fail_recovery_execution.sql
	failRecoveryExecutionSQL string
	//go:embed sql/platform/complete_recovery_set.sql
	completeRecoverySetSQL string
)
