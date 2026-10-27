package application

import "github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"

type RestoreReceiptPayload = recoveryset.RestoreReceiptPayload

func ParseRecoveryReceipt(document []byte) (RestoreReceiptPayload, error) {
	return recoveryset.ParseRestoreReceipt(document)
}
