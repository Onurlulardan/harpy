package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5"
)

func executeFinalizerIssuance(
	ctx context.Context, config invocation,
) (commandResult, error) {
	profile := *config.IssueFinalizer
	now := time.Now().UTC()
	prerequisites, err := loadFinalizationPrerequisites(
		profile.finalizationPrerequisitePaths, now,
	)
	if err != nil {
		return commandResult{}, err
	}
	signer, err := loadAuthoritySigner(profile.SigningKeyFile, prerequisites.authority.root)
	if err != nil {
		return commandResult{}, err
	}
	connection, err := openAuthorityDatabase(
		ctx, profile.ClusterAdminDSNFile, config.Environment,
	)
	if err != nil {
		return commandResult{}, err
	}
	defer connection.Close(context.Background())
	if err := preflightClusterAdmin(
		ctx, connection, prerequisites.adoption.Target(), now,
	); err != nil {
		return commandResult{}, err
	}
	finalizerConnection, err := openFinalizerDatabase(
		ctx, profile.FinalizerDSNFile, config.Environment,
	)
	if err != nil {
		return commandResult{}, err
	}
	defer finalizerConnection.Close(context.Background())
	connectedFinalizerValidUntil, err := preflightFinalizer(
		ctx, finalizerConnection, prerequisites.adoption.Target(), now,
	)
	if err != nil {
		return commandResult{}, err
	}
	transaction, err := connection.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return commandResult{}, fmt.Errorf("finalizer issuance transaction failed")
	}
	defer transaction.Rollback(context.Background())
	if err := recordCredentialReceipts(ctx, transaction, prerequisites); err != nil {
		return commandResult{}, err
	}
	observation, err := observeFinalizerAuthority(
		ctx, transaction, prerequisites.adoption.OperationID(),
		prerequisites.adoption.RunID(),
	)
	if err != nil {
		return commandResult{}, err
	}
	if err := observation.validate(
		prerequisites.adoption.Target(), prerequisites.terminal.value.TerminalState,
		connectedFinalizerValidUntil,
	); err != nil {
		return commandResult{}, err
	}
	signed, err := buildSignedFinalizerLease(
		profile.FinalizerLeaseID, prerequisites, observation, signer,
	)
	if err != nil {
		return commandResult{}, err
	}
	if err := verifyIssuedFinalizerLease(signed, prerequisites, observation.databaseTime); err != nil {
		return commandResult{}, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return commandResult{}, fmt.Errorf("finalizer issuance commit failed")
	}
	replayed, err := publishIssuedFinalizerLease(profile, signed)
	if err != nil {
		return commandResult{}, err
	}
	return commandResult{Action: "issue_finalizer",
		OperationID: prerequisites.adoption.OperationID(), EvidenceSHA256: signed.digest,
		Replayed: replayed}, nil
}

func verifyIssuedFinalizerLease(
	signed signedFinalizerLease, prerequisites verifiedFinalizationPrerequisites,
	now time.Time,
) error {
	verified, err := attestation.VerifyCutoverFinalizerLease(
		signed.document, signed.signatureDocument, prerequisites.adoption,
		prerequisites.cleanup, prerequisites.receipt, prerequisites.authority.root, now,
	)
	if err != nil {
		return fmt.Errorf("issued finalizer lease self-verification failed")
	}
	if _, err := attestation.VerifyCutoverTerminalEvidence(
		prerequisites.terminal.document, prerequisites.adoption, verified,
	); err != nil {
		return fmt.Errorf("issued terminal evidence self-verification failed")
	}
	return nil
}

func publishIssuedFinalizerLease(
	profile finalizerIssueInvocation, signed signedFinalizerLease,
) (bool, error) {
	replayed := existingArtifactMatches(profile.FinalizerLeasePath, signed.document) &&
		existingArtifactMatches(
			profile.FinalizerLeaseSignaturePath, signed.signatureDocument,
		)
	if err := publishNoReplace(profile.FinalizerLeasePath, signed.document); err != nil {
		return false, fmt.Errorf("finalizer lease publication failed")
	}
	if err := publishNoReplace(
		profile.FinalizerLeaseSignaturePath, signed.signatureDocument,
	); err != nil {
		return false, fmt.Errorf("finalizer lease signature publication failed")
	}
	return replayed, nil
}
