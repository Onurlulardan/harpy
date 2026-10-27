package licenseoffering

import _ "embed"

//go:embed sql/platform/publish_license_offering.sql
var publishLicenseOfferingSQL string
