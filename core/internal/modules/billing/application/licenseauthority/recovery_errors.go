package licenseauthority

import "errors"

var (
	ErrRestoredLicenseConfiguration = errors.New("restored license coordinator is unavailable")
	ErrRestoredLicenseScope         = errors.New("restored license scope is invalid")
	ErrRestoredLicenseReceipt       = errors.New("restored license receipt is invalid")
)
