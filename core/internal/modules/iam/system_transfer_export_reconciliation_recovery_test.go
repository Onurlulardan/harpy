package iam

import "testing"

func TestOpenSystemTransferExportReconciliationRecoveryRejectsNilCoordinator(
	t *testing.T,
) {
	recovery, err := OpenSystemTransferExportReconciliationRecovery(nil)
	if err == nil || recovery != nil {
		t.Fatal("nil coordinator must reject export reconciliation recovery")
	}
}

func TestSystemTransferExportReconciliationRecoveryRejectsNilReceiver(
	t *testing.T,
) {
	var recovery *systemTransferExportReconciliationRecovery
	if err := recovery.RecoverExpiredSystemTransferExports(nil, [16]byte{}, 0); err == nil {
		t.Fatal("nil recovery must fail closed")
	}
}
