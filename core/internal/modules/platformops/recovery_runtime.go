package platformops

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/backuparchive"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/pgtools"
	platformopspostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres"
	executionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/recoveryexecution"
	copypostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/recoveryexecution/copy"
	snapshotpostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/recoveryexecution/snapshot"
	verifypostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/recoveryexecution/verify"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/recoverys3"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryexecution"
	recoverycopy "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryexecution/copy"
	recoverysnapshot "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryexecution/snapshot"
	recoveryverify "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryexecution/verify"
	recovery "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryorchestration"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
)

type RecoveryFactories = recovery.RecoveryFactories
type RestoreFailureKind = recovery.RestoreFailureKind

type RecoveryExecutor interface {
	Run(context.Context, *admission.Gate) error
	Close()
}

type backupRecoveryExecutor struct {
	executor recovery.BackupExecutor
}

const (
	RestoreFailureInvalid    = recovery.RestoreFailureInvalid
	RestoreFailureDependency = recovery.RestoreFailureDependency
	RestoreFailureResumable  = recovery.RestoreFailureResumable
	RestoreFailureTerminal   = recovery.RestoreFailureTerminal
)

func NewDatabaseBackupExecutor(
	ctx context.Context,
	database any,
	settings config.DatabaseBackupExecutor,
	factories RecoveryFactories,
) (RecoveryExecutor, error) {
	if !factories.Valid() {
		return nil, fmt.Errorf("database backup module factories are unavailable")
	}
	owned, ok := database.(interface{ RuntimeOwnerDatabase() })
	if !ok {
		return nil, fmt.Errorf("database backup pool is unavailable")
	}
	runtime := settings.LongRunning().Runtime()
	trustDocument, err := readRecoveryRuntimeFile(
		settings.RecoveryTrustBundle().Path().Value(),
		recoverysigningv1.MaximumBundleBytes, false,
	)
	if err != nil {
		return nil, err
	}
	trust, err := recoverysigningv1.Parse(trustDocument)
	if err != nil {
		return nil, err
	}
	seed, err := readRecoveryRuntimeFile(
		settings.ManifestSigningKey().Path().Value(), ed25519.SeedSize, true,
	)
	if err != nil || len(seed) != ed25519.SeedSize {
		clear(seed)
		return nil, fmt.Errorf("recovery signing key is invalid")
	}
	signer := ed25519.NewKeyFromSeed(seed)
	clear(seed)
	storage, err := recoverys3.LoadDatabaseBackupStorage(
		ctx, settings.RecoveryStorageManifest().Path().Value(),
		runtime.Environment().String(), readRecoveryRuntimeFile,
	)
	if err != nil {
		clear(signer)
		return nil, err
	}
	targets, err := factories.Targets(database, "backup")
	if err != nil {
		clear(signer)
		return nil, err
	}
	target, err := targets.Get(ctx)
	if err != nil {
		clear(signer)
		return nil, err
	}
	documents, err := factories.Documents(database, documentsapi.RecoveryIdentity{
		EnvironmentID: target.EnvironmentID(), TenantID: target.TenantID(),
		TenantSetSHA256: target.TenantSetSHA256(),
	})
	if err != nil {
		clear(signer)
		return nil, err
	}
	executionStore, err := executionpostgres.Bind(owned)
	if err != nil {
		clear(signer)
		return nil, err
	}
	execution, err := recoveryexecution.New(executionStore)
	snapshotStore, snapshotErr := snapshotpostgres.Bind(owned)
	copyStore, copyErr := copypostgres.Bind(owned)
	verifyStore, verifyErr := verifypostgres.Bind(owned)
	iamEvidence, iamErr := factories.IAMEvidence(database)
	authorityEvidence, authorityErr := factories.AuthorityEvidence(database)
	if err != nil || snapshotErr != nil || copyErr != nil || verifyErr != nil ||
		iamErr != nil || authorityErr != nil {
		clear(signer)
		return nil, fmt.Errorf("database backup owner bindings are unavailable")
	}
	tools := pgtools.New(
		pgtools.OSRunner{}, "/usr/bin/pg_dump", "/usr/bin/pg_restore",
	)
	archive := backuparchive.New(tools, tools)
	executor, err := recovery.NewBackupExecutor(recovery.BackupDependencies{
		Execution: execution, Snapshot: recoverysnapshot.New(snapshotStore),
		Copy: recoverycopy.New(copyStore), Verify: recoveryverify.New(verifyStore),
		BeginSnapshot: func(work context.Context) (recovery.RecoverySnapshot, error) {
			return platformopspostgres.BeginRecoverySnapshot(work, database)
		},
		NewAttemptID: func(now time.Time) (string, error) {
			value, idErr := randomid.NewUUIDv7(now)
			return value.String(), idErr
		},
		Documents: documents, IAMEvidence: iamEvidence,
		AuthorityEvidence: authorityEvidence, Target: target,
		Storage: storage, Archive: archive, Trust: trust, Signer: signer,
		DSN: runtime.Postgres().DSN().Value(),
	})
	if err != nil {
		clear(signer)
		return nil, err
	}
	return &backupRecoveryExecutor{executor: executor}, nil
}

func (executor *backupRecoveryExecutor) Run(
	ctx context.Context,
	gate *admission.Gate,
) error {
	if executor == nil || executor.executor == nil || gate == nil {
		return fmt.Errorf("database backup executor is unavailable")
	}
	return executor.executor.Run(
		ctx,
		func() bool {
			permit, admitted := gate.Admit()
			return admitted && permit.Granted()
		},
		gate.Draining,
	)
}

func (executor *backupRecoveryExecutor) Close() {
	if executor != nil && executor.executor != nil {
		executor.executor.Close()
	}
}

func ClassifyDatabaseRestoreError(err error) RestoreFailureKind {
	return recovery.ClassifyDatabaseRestoreError(err)
}
