package operator

import (
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformopsmodule "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
)

type restoredIdentityArtifacts struct {
	receiptDocument  []byte
	receipt          platformopsmodule.RestoreReceiptPayload
	restoreOperation iammodule.UUID
	environment      iammodule.UUID
	recoverySet      iammodule.UUID
	security         iammodule.RestoredSecurityEvidence
	tenant           iammodule.UUID
	owner            iammodule.UUID
	anchor           iammodule.UUID
	newDatabase      iammodule.UUID
	protectedHead    [32]byte
	receiptSHA256    [32]byte
	custodyDocument  []byte
	custody          iammodule.RestoreCustodyManifest
	occurredAt       time.Time
}

type identityReconcileOutcome struct {
	OperationID        iammodule.UUID
	EnvironmentID      iammodule.UUID
	TenantID           iammodule.UUID
	State              string
	Replayed           bool
	InvalidationSHA256 [32]byte
	CustodyGeneration  int64
	CustodySHA256      [32]byte
}
