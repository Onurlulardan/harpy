package restoredlicense

import _ "embed"

//go:embed sql/platform/mint_restored_license.sql
var mintRestoredLicenseGrantSQL string
