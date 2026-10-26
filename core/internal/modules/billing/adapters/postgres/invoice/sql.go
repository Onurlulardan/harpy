package invoice

import _ "embed"

var (
	//go:embed sql/platform/resolve_authority.sql
	resolveAuthoritySQL string
	//go:embed sql/platform/assert_authority.sql
	assertAuthoritySQL string
	//go:embed sql/platform/lock_draft.sql
	lockDraftSQL string
	//go:embed sql/platform/find_finalization.sql
	findFinalizationSQL string
	//go:embed sql/platform/write_finalization.sql
	writeFinalizationSQL string
)
