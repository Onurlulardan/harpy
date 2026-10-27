package recoveryorchestration

import (
	"bytes"
	"context"
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (executor *databaseBackupExecutor) finalize(
	ctx context.Context,
	claim ports.RecoveryExecutionClaim,
	identity ports.RecoveryExecutionIdentity,
	components map[string]databaseBackupComponent,
	documentsDigest string,
) error {
	bound, err := recoveryComponents(components)
	if err != nil {
		return err
	}
	createdAt := claim.DatabaseNow.UTC().Truncate(time.Microsecond)
	payload := recoveryset.ManifestPayload{
		Schema: recoveryset.ManifestSchema, Portability: recoveryset.SameEnvironment,
		Identity: recoveryset.EnvironmentIdentity{
			EnvironmentID:          claim.EnvironmentID,
			DeploymentMode:         recoveryset.ModeOnPremise,
			PrimaryBindingID:       claim.PrimaryBindingID,
			TenantSetSHA256:        claim.TenantSetSHA256,
			OwnerBindingID:         claim.OwnerBindingID,
			OwnerBindingGeneration: claim.OwnerBindingGeneration,
			DatabaseLineageID:      claim.DatabaseLineageID,
		},
		DatabaseSystemID: claim.DatabaseSystemID, Components: bound,
		DatabaseArchiveSHA256:     bound.DatabaseArchive.CiphertextSHA256,
		DocumentsMerkleSHA256:     documentsDigest,
		KeyCustodySHA256:          bound.KeyCustody.CiphertextSHA256,
		OwnerAnchorSHA256:         bound.OwnerAnchor.CiphertextSHA256,
		ProtectedHeadAckSHA256:    bound.ProtectedAuthority.CiphertextSHA256,
		RecoveryTrustBundleSHA256: claim.RecoveryTrustBundleSHA256,
		RecoverySigningGeneration: claim.SigningGeneration,
		CreatedAt:                 createdAt,
	}
	payloadBytes, err := payload.CanonicalBytes()
	if err != nil {
		return err
	}
	artifact, err := environmentrecoveryv1.Build(
		environmentrecoveryv1.Input{
			Purpose:       environmentrecoveryv1.PurposeRecoveryManifest,
			EnvironmentID: claim.EnvironmentID, OperationID: claim.OperationID,
			ArtifactID: claim.RecoverySetID, Sequence: claim.Sequence,
			IssuedAt: createdAt,
			ExpiresAt: createdAt.Add(time.Duration(
				executor.storage.RetentionDays(),
			) * 24 * time.Hour),
			Payload: payloadBytes,
		},
		executor.trust, executor.signer,
	)
	if err != nil {
		return err
	}
	prefix, err := executor.storage.Prefix(
		claim.EnvironmentID, claim.RecoverySetID,
	)
	if err != nil {
		return err
	}
	document := artifact.CanonicalBytes()
	if _, err = executor.storage.Writer(claim.DatabaseNow).PutVerified(
		ctx, ports.WriteRequest{
			Key: prefix + "manifest.json", Body: bytes.NewReader(document),
			ContentLength:  int64(len(document)),
			ExpectedSHA256: artifact.SHA256(), ExpectedBytes: int64(len(document)),
		},
	); err != nil {
		return err
	}
	if err = executor.execution.Complete(
		ctx, ports.RecoveryCompletionRecord{
			RecoverySetID: claim.RecoverySetID, AttemptID: identity.AttemptID,
			LeaseToken: identity.LeaseToken, ManifestSHA256: artifact.SHA256(),
			ManifestBytes: document, DatabaseSystemID: claim.DatabaseSystemID,
		},
	); err != nil {
		return err
	}
	return executor.documents.ReleaseHold(
		ctx, claim.RecoverySetID, decodeDigest32(documentsDigest),
	)
}
