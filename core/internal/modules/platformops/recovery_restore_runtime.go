package platformops

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	documentsapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/backuparchive"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/pgtools"
	platformopspostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres"
	restorepostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/postgres/restoreexecution"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/recoverys3"
	recovery "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryorchestration"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/restoreexecution"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func RunDatabaseRestore(
	ctx context.Context,
	database any,
	settings config.DatabaseRestore,
	_ schemacontract.EmbeddedContract,
	factories RecoveryFactories,
	openScratch func(context.Context) (any, error),
) error {
	if !factories.Valid() || openScratch == nil {
		return fmt.Errorf("database restore module factories are unavailable")
	}
	owned, ok := database.(interface{ RuntimeOwnerDatabase() })
	if !ok {
		return fmt.Errorf("database restore pool is unavailable")
	}
	store, err := restorepostgres.Bind(owned)
	if err != nil {
		return err
	}
	execution := restoreexecution.New(store)
	if execution == nil {
		return fmt.Errorf("restore execution service is unavailable")
	}
	runtime := settings.Runtime()
	artifacts := restoreArtifacts(settings.Artifacts())
	storage, err := recoverys3.LoadDatabaseRestoreStorage(
		ctx, artifacts.RecoveryStorageManifest,
		runtime.Environment().String(), readRecoveryRuntimeFile,
	)
	if err != nil {
		return err
	}
	tools := pgtools.New(
		pgtools.OSRunner{}, "/usr/bin/pg_dump", "/usr/bin/pg_restore",
	)
	return recovery.RunRestore(ctx, recovery.RestoreDependencies{
		Execution: execution, Storage: storage,
		Archive: backuparchive.New(tools, tools),
		OpenScratch: func(
			work context.Context,
			environmentID, tenantID string,
			tenantSetSHA256 [32]byte,
			databaseSystemID string,
		) (recovery.RestoreScratch, error) {
			raw, openErr := openScratch(work)
			if openErr != nil {
				return nil, openErr
			}
			return bindRestoreScratch(
				work, raw, factories, environmentID, tenantID,
				tenantSetSHA256, databaseSystemID,
			)
		},
		NewAttemptID: func(now time.Time) (string, error) {
			value, idErr := randomid.NewUUIDv7(now)
			return value.String(), idErr
		},
		NewDatabaseID: func() (string, error) {
			value, idErr := randomid.NewUUIDv4()
			return value, idErr
		},
		LocatorSHA256: pgtools.LocatorSHA256,
		Artifacts:     artifacts, Environment: runtime.Environment().String(),
		ControlDSN: runtime.ControlPostgres().DSN().Value(),
		ScratchDSN: runtime.ScratchPostgres().DSN().Value(),
	})
}

type restoreScratch struct {
	database  interface{ Close() }
	documents documentsapi.Recovery
	importer  *restoreexecution.Service
}

func (scratch *restoreScratch) Close() {
	if scratch != nil && scratch.database != nil {
		scratch.database.Close()
	}
}

func (scratch *restoreScratch) Documents() documentsapi.Recovery {
	if scratch == nil {
		return nil
	}
	return scratch.documents
}

func (scratch *restoreScratch) ImportReceipt(
	ctx context.Context,
	record ports.RestoreReceiptImport,
) error {
	if scratch == nil || scratch.importer == nil {
		return fmt.Errorf("restore receipt importer is unavailable")
	}
	return scratch.importer.ImportReceipt(ctx, record)
}

func bindRestoreScratch(
	ctx context.Context,
	database any,
	factories RecoveryFactories,
	environmentID, tenantID string,
	tenantSetSHA256 [32]byte,
	databaseSystemID string,
) (recovery.RestoreScratch, error) {
	owned, ownerOK := database.(interface{ RuntimeOwnerDatabase() })
	closer, closeOK := database.(interface{ Close() })
	if !ownerOK || !closeOK {
		return nil, fmt.Errorf("restore scratch database is unavailable")
	}
	fail := func(err error) (recovery.RestoreScratch, error) {
		closer.Close()
		return nil, err
	}
	identityStore, err := platformopspostgres.NewStoreForRestore(owned)
	if err != nil {
		return fail(err)
	}
	current, err := identityStore.CurrentIdentity(ctx, environmentID)
	if err != nil || current.Identity.EnvironmentID != environmentID ||
		current.Identity.PrimaryBindingID != tenantID ||
		current.Identity.TenantSetSHA256 != hex.EncodeToString(tenantSetSHA256[:]) ||
		current.DatabaseSystemID != databaseSystemID {
		return fail(fmt.Errorf("restored database identity differs"))
	}
	targets, err := factories.Targets(database, "restore")
	if err != nil {
		return fail(err)
	}
	target, err := targets.Get(ctx)
	if err != nil || target.EnvironmentID() != environmentID ||
		target.TenantID() != tenantID ||
		target.TenantSetSHA256() != tenantSetSHA256 {
		return fail(fmt.Errorf("restored tenant identity differs"))
	}
	documents, err := factories.Documents(
		database, documentsapi.RecoveryIdentity{
			EnvironmentID: environmentID, TenantID: tenantID,
			TenantSetSHA256: tenantSetSHA256,
		},
	)
	if err != nil {
		return fail(err)
	}
	executionStore, err := restorepostgres.Bind(owned)
	if err != nil {
		return fail(err)
	}
	return &restoreScratch{
		database: closer, documents: documents,
		importer: restoreexecution.New(executionStore),
	}, nil
}

func restoreArtifacts(
	artifacts config.DatabaseRestoreArtifacts,
) recovery.RestoreArtifacts {
	branch := recovery.RestoreExistingReadable
	if artifacts.Branch() == config.RestoreRehydrate {
		branch = recovery.RestoreRehydrate
	}
	return recovery.RestoreArtifacts{
		Branch:                  branch,
		RecoveryStorageManifest: artifacts.RecoveryStorageManifest().Path().Value(),
		RecoveryTrustBundle:     artifacts.RecoveryTrustBundle().Path().Value(),
		OrchestratorTrustBundle: artifacts.OrchestratorTrustBundle().Path().Value(),
		ProgressPath:            artifacts.ProgressPath().Path().Value(),
		DrainAckPath:            artifacts.DrainAckPath().Path().Value(),
		SwitchAckPath:           artifacts.SwitchAckPath().Path().Value(),
		ReceiptPath:             artifacts.ReceiptPath().Path().Value(),
		RestorePlan:             artifacts.RestorePlan().Path().Value(),
		DRKEKKeyring:            artifacts.DRKEKKeyring().Path().Value(),
		ReceiptSigningKey:       artifacts.ReceiptSigningKey().Path().Value(),
		RehydrateRequest:        artifacts.RehydrateRequest().Path().Value(),
		RecoveryKit:             artifacts.RecoveryKit().Path().Value(),
	}
}
