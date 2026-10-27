package tenantsettingspostgres

import _ "embed"

var (
	//go:embed sql/platform/read.sql
	readSQL string
	//go:embed sql/platform/update.sql
	updateSQL string
)
