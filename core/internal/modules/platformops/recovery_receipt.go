package platformops

import platformopsapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application"

// RestoreReceiptPayload is the module-owned restore receipt contract exposed to
// composition without leaking the recoveryset domain package.
type RestoreReceiptPayload = platformopsapp.RestoreReceiptPayload

func ParseRestoreReceipt(document []byte) (RestoreReceiptPayload, error) {
	return platformopsapp.ParseRecoveryReceipt(document)
}
