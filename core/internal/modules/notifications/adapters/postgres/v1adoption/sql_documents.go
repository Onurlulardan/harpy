package v1adoption

import _ "embed"

var (
	//go:embed sql/tenant/source_exists.sql
	sourceExistsSQL string
	//go:embed sql/tenant/read_source_page.sql
	readSourcePageSQL string
	//go:embed sql/tenant/insert_run.sql
	insertRunSQL string
	//go:embed sql/tenant/lock_run.sql
	lockRunSQL string
	//go:embed sql/tenant/begin_applying.sql
	beginApplyingSQL string
	//go:embed sql/tenant/get_adopted.sql
	getAdoptedSQL string
	//go:embed sql/tenant/upsert_counter.sql
	upsertCounterSQL string
	//go:embed sql/tenant/insert_message.sql
	insertMessageSQL string
	//go:embed sql/tenant/insert_effect.sql
	insertEffectSQL string
	//go:embed sql/tenant/insert_delivery.sql
	insertDeliverySQL string
	//go:embed sql/tenant/insert_attempt.sql
	insertAttemptSQL string
	//go:embed sql/tenant/insert_issue.sql
	insertIssueSQL string
	//go:embed sql/tenant/save_checkpoint.sql
	saveCheckpointSQL string
	//go:embed sql/tenant/complete_run.sql
	completeRunSQL string
)
