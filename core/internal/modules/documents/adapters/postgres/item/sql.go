package item

import _ "embed"

var (
	//go:embed sql/tenant/get.sql
	getSQL string
	//go:embed sql/tenant/list.sql
	listSQL string
	//go:embed sql/tenant/list_versions.sql
	listVersionsSQL string
	//go:embed sql/tenant/update.sql
	updateSQL string
	//go:embed sql/tenant/move.sql
	moveSQL string
	//go:embed sql/tenant/trash.sql
	trashSQL string
	//go:embed sql/tenant/restore.sql
	restoreSQL string
	//go:embed sql/tenant/request_purge.sql
	requestPurgeSQL string
	//go:embed sql/tenant/cancel_purge.sql
	cancelPurgeSQL string
)
