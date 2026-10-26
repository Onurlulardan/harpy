package uploadreconcile

import _ "embed"

var (
	//go:embed sql/tenant/prepare.sql
	prepareSQL string
	//go:embed sql/tenant/finalize.sql
	finalizeSQL string
)
