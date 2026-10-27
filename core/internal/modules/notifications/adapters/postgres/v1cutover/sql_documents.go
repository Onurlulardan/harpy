package v1cutover

import _ "embed"

var (
	//go:embed sql/tenant/get_authority.sql
	getAuthoritySQL string
	//go:embed sql/tenant/bootstrap.sql
	bootstrapSQL string
	//go:embed sql/tenant/quiesce.sql
	quiesceSQL string
	//go:embed sql/tenant/record_evidence.sql
	recordEvidenceSQL string
	//go:embed sql/tenant/fence.sql
	fenceSQL string
	//go:embed sql/tenant/activate_v2.sql
	activateV2SQL string
)
