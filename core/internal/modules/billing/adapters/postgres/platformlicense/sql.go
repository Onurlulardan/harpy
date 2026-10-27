package platformlicense

import _ "embed"

var (
	//go:embed sql/platform/read_license.sql
	readLicenseSQL string
	//go:embed sql/platform/request_refresh.sql
	requestRefreshSQL string
	//go:embed sql/platform/read_refresh_operation.sql
	readRefreshOperationSQL string
)
