package iam

import (
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func TestSystemExportClaimBindsExactExternalJob(t *testing.T) {
	claim := canonicalExportReconcileClaim(t, publicresource.Users)
	claim.kind = SystemExportJobKind
	claim.maximum = systemtransfer.SystemExportMaximumAttempts
	claim.effectKey = systemExportEffectPrefix +
		claim.effectKey[len(systemExportReconcileEffectPrefix):]
	decoded, err := decodeSystemTransferExportClaim(claim, SystemExportJobPolicy())
	if err != nil || decoded.ExportID.String() !=
		claim.effectKey[len(systemExportEffectPrefix):] ||
		decoded.AttemptNumber != 1 ||
		decoded.MaximumAttempts != systemtransfer.SystemExportMaximumAttempts {
		t.Fatalf("decoded=%#v err=%v", decoded, err)
	}
}

func TestSystemExportClaimRejectsReconcileKindAndEffectDrift(t *testing.T) {
	claim := canonicalExportReconcileClaim(t, publicresource.Roles)
	claim.maximum = systemtransfer.SystemExportMaximumAttempts
	if _, err := decodeSystemTransferExportClaim(claim, SystemExportJobPolicy()); err == nil {
		t.Fatal("reconciliation job was accepted as an export")
	}
	claim.kind = SystemExportJobKind
	claim.effectKey = systemExportEffectPrefix +
		claim.effectKey[len(systemExportReconcileEffectPrefix):] + "-drift"
	if _, err := decodeSystemTransferExportClaim(claim, SystemExportJobPolicy()); err == nil {
		t.Fatal("drifted export effect key was accepted")
	}
}
