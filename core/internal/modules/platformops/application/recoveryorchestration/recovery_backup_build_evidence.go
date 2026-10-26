package recoveryorchestration

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverykeymaterial"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseBackupExecutor) buildEvidenceComponents(
	ctx context.Context,
	claim ports.RecoveryExecutionClaim,
	prefix string,
	writer RecoveryObjectWriter,
	material recoverykeymaterial.Material,
	keys backupcrypto.SealKeys,
) (map[string]databaseBackupComponent, error) {
	custody, err := buildKeyCustodyEvidence(
		claim.EnvironmentID, claim.RecoverySetID, material,
	)
	if err != nil {
		return nil, err
	}
	anchors, err := executor.iam.ReadOwnerAnchors(ctx, claim.EnvironmentID)
	if err != nil {
		return nil, err
	}
	owner, err := buildOwnerAnchorEvidence(
		claim.EnvironmentID, claim.RecoverySetID,
		claim.OwnerBindingID, claim.OwnerBindingGeneration, anchors,
	)
	if err != nil {
		return nil, err
	}
	head, err := executor.authority.ReadRecoveryHead(ctx, claim.EnvironmentID)
	if err != nil {
		return nil, err
	}
	authority, err := buildProtectedAuthorityEvidence(
		claim.EnvironmentID, claim.RecoverySetID,
		claim.RecoveryTrustBundleSHA256, head,
	)
	if err != nil {
		return nil, err
	}
	documents := map[string][]byte{
		"key_custody": custody, "owner_anchor": owner,
		"protected_authority": authority,
	}
	result := make(map[string]databaseBackupComponent, len(documents))
	for kind, document := range documents {
		component, writeErr := putEncryptedDocument(
			ctx, writer, componentKey(prefix, kind), document,
			archiveContext(claim, kind,
				material.DRKeyID(), material.RecoveryKeyID()), keys,
		)
		clear(document)
		if writeErr != nil {
			return nil, writeErr
		}
		component.kind = kind
		result[kind] = component
	}
	return result, nil
}
