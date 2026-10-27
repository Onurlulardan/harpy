package compatibility

import _ "embed"

var (
	//go:embed sql/tenant/record_item.sql
	recordItemSQL string
	//go:embed sql/tenant/accept.sql
	acceptSQL string
	//go:embed sql/platform/resolve.sql
	resolveSQL string
	//go:embed sql/platform/record_hit.sql
	recordHitSQL string
	//go:embed sql/platform/acknowledge_consumers.sql
	acknowledgeConsumersSQL string
	//go:embed sql/platform/retire.sql
	retireSQL string
)
