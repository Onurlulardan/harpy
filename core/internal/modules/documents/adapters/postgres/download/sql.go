package download

import _ "embed"

var (
	//go:embed sql/tenant/get_current.sql
	getCurrentSQL string
	//go:embed sql/tenant/get_version.sql
	getVersionSQL string
)
